const mongoose = require("mongoose");
const moment = require("moment");
const twine = require("twine-action-lib-service");
require('moment-round');

module.exports["ai.twine.appointment.search"] = function (ctx, req) {
  twine.registerModels(ctx);

  let serviceKey = req.slots["found_service_data"];
  if (!serviceKey) {
    throw new Error("Missing service key");
  }

  const businessId = serviceKey.businessId;
  if (!businessId) {
    throw new Error("Missing business ID");
  }

  const calendarId = new mongoose.Types.ObjectId("5a36db6b4c8d1b000547253e");

  const calendar = new Calendar(businessId, calendarId, ctx.database, ctx.logger, ctx.models);

  // serviceKey = {
  //   module: "ai.twine.foo",
  //   businessId: "test",
  //   _id: new mongoose.Types.ObjectId(),
  // };

  let service = null;
  let appointment = null;

  let time = moment();
  let isExactSearch = false;
  if (req.slots.time) {
    console.log("utc offset: %s", moment().utcOffset())
    time = moment(req.slots.time).subtract(moment().utcOffset(), "minutes");

    isExactSearch = (time.hour() > 0 || time.minute() > 0);
  }

  return calendar.findService({ _id: serviceKey.id, businessId: businessId })
    .then((service) => {

      ctx.logger.info(`Looking for appointment ${time}`);
      time.ceil(15, "minutes");
      ctx.logger.info(`Looking for appointment ${time}`);

      let duration = moment.duration(3600, "s");
      if (service.duration) {
        duration = moment.duration(service.duration, "s");
      }

      ctx.logger.debug(`Service ${service.id} duration: ${duration}`);

      return calendar.findPotentialAppointment(time, duration, service, isExactSearch);
    })
    .then((entry) => {
      if (entry) {
        ctx.logger.debug(`Found potential appointment ${entry.appt}`);
        ctx.setSlot("found_appointment_match_type", entry.isExact ? "exact" : "suggested");
        ctx.setSlot("found_appointment_data", {
          businessId: businessId,
          id: entry.appt.id,
          calendarId: entry.appt.calendarId.toString(),
        });
      } else {
        ctx.logger.debug(`No potential appointments found for ${moment(time).startOf("day")}`);
        ctx.setSlot("found_appointment_match_type", null);
        ctx.setSlot("found_appointment_data", null);
      }
    });
}

class Calendar {
  constructor(businessId, calendarId, db, logger, models) {
    this.db = db;
    this.logger = logger;
    this.businessId = businessId;
    this.calendarId = calendarId;
    this.models = models;
  }


  findService(serviceKey) {
    return this.models.Service
      .findOne(serviceKey)
      .then((service) => {
        if (!service) {
          throw new Error("service not found");
        }

        this.logger.debug(`Found service: ${service}`);
        return service;
      });
  }

  findAppointmentInSlot(startTime, duration) {
    const endTime = moment(startTime).add(duration);

    return this.models.Appointment
      .findOne({
        businessId: this.businessId,
        calendarId: this.calendarId,
        startsAt: {
          $gte: startTime,
          $lt: endTime
        },
      });
  }

  findPotentialAppointment(initialTime, duration, service, isExact = true) {
    let time = moment(initialTime).ceil(15, "minutes");

    this.logger.debug("Looking for appointment starting at: %s", time.inspect());

    if (time.isBefore(moment().add(30, "minutes"))) {
      this.logger.debug("  -> Before current time, moving to now +30min");
      let newStart = moment().add(30, "minutes");
      return this.findPotentialAppointment(newStart, duration, service, false);
    }

    if (!this.isWithinBusinessHours(time, duration)) {
      let startOfDay = moment(time).startOf("day").add(9, "hours");

      // If we're not within business hours because of starting too early, start the search at the beginning
      // of the day and go from there.
      if (time.isBefore(startOfDay)) {
        this.logger.debug("  -> Before start of business, moving to start of day");
        return this.findPotentialAppointment(startOfDay, duration, service, false);
      }

      this.logger.debug("  -> outside of business hours");

      return Promise.resolve();
    }

    return this.findAppointmentInSlot(time, duration)
      .then((appt) => {
        if (!appt) {
          return this.createPendingAppointment(time, duration, service, isExact)
            .then((created) => {
              return {
                appt: created,
                isExact: isExact,
              };
            });
        } else {
          return this.findPotentialAppointment(moment(appt.endsAt), duration, service, false);
        }
      });
  }

  isWithinBusinessHours(startsAt, duration) {
    let startOfDay = moment(startsAt).startOf("day").add(9, "hours");
    let endOfDay = moment(startOfDay).add(8, "hours");
    let endsAt = moment(startsAt).add(duration);

    return startsAt.isBetween(startOfDay, endOfDay, null, "[)") &&
           endsAt.isBetween(startOfDay, endOfDay, null, "(]");
  }

  createPendingAppointment(startsAt, duration, service, isExact) {
    let endsAt = moment(startsAt).add(duration);

    let appt = new this.models.Appointment({
      businessId: this.businessId,
      calendarId: this.calendarId,
      serviceId: service.id,
      name: service.title,
      startsAt: startsAt,
      endsAt: endsAt,
      pending: !isExact,
      updatedAt: moment(),
    });

    return appt.save();
  }
}
