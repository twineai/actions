const moment = require("moment");
const twine = require("twine-action-lib-service");

module.exports["ai.twine.appointment.accept_proposal"] = function (ctx, req) {
  twine.registerModels(ctx);
  let accepter = new Accepter(ctx.database, ctx.logger, ctx.models);

  let appointmentKey = req.slots["found_appointment_data"];
  if (!appointmentKey) {
    throw new Error("Missing appointment key");
  }

  return accepter
    .findAppointment(appointmentKey)
    .then((appointment) => {
      const time = moment(appointment.startsAt);
      const dateTimeString = time.calendar();
      ctx.speak(`We're available ${dateTimeString}.`, true);
    });
};

class Accepter {
  constructor(db, logger, models) {
    this.db = db;
    this.logger = logger;
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

  findAppointment(appointmentKey) {
    return this.models.Appointment
      .findOne(appointmentKey)
      .then((appointment) => {
        if (!appointment) {
          throw new Error("appointment not found");
        }

        this.logger.debug(`Found appointment: ${appointment}`);
        return appointment;
      });
  }
}
