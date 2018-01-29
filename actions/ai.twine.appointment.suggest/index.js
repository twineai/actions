const moment = require("moment");
const twine = require("twine-action-lib-service");
const util = require("util");

var suggestionIdx = 0;

module.exports["ai.twine.appointment.suggest"] = function (ctx, req) {
  twine.registerModels(ctx);

  let suggester = new Suggester(ctx.database, ctx.logger, ctx.models);

  let appointmentKey = req.slots["found_appointment_data"];
  if (!appointmentKey) {
    throw new Error("Missing appointment key");
  }

  return suggester
    .findAppointment(appointmentKey)
    .then((appointment) => {
      return Promise.all([
        appointment,
        suggester.findService({ businessId: appointmentKey.businessId, _id: appointment.serviceId }),
      ]);
    })
    .then(([appointment, service]) => {
      const time = moment(appointment.startsAt);
      const dateTimeString = time.calendar();

      const suggestionTemplates = [
        `I have an opening for a ${service.title} ${dateTimeString}. Will that work?`,
        `Can you do ${dateTimeString}? I have an opening then for a ${service.title}.`,
      ];

      const text = suggestionTemplates[suggestionIdx];

      if (suggestionIdx == 0) {
        suggestionIdx = 1;
      } else {
        suggestionIdx = 0;
      }

      ctx.speak(text, true);
    });
};

class Suggester {
  constructor(db, logger, models) {
    this.db = db;
    this.logger = logger;
    this.models = models;
  }

  findService(serviceKey) {
    return this.models.Service
      .findOne(serviceKey)
      .exec()
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
      .exec()
      .then((appointment) => {
        if (!appointment) {
          throw new Error("appointment not found");
        }

        this.logger.debug(`Found appointment: ${appointment}`);
        return appointment;
      });
  }
}
