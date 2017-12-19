const mongoose = require("mongoose");
const moment = require("moment");
const twine = require("twine-action-lib-service");
require('moment-round');

module.exports["ai.twine.appointment.suggest_dates"] = function (ctx, req) {
  let time = moment();
  if (req.slots.time) {
    time = moment(req.slots.time).subtract(moment().utcOffset(), "minutes");
  }

  if (req.slots.found_appointment_data) {
    // Have appointment data, means we've suggested something and they don't want it, provide a more
    // generic response.
    ctx.speak("ai.twine.appointment.request_different_date");
  } else {
    if (isWithinBusinessHours(time)) {
      // No appointment data == searched for a time and couldn't get it.
      const dateString = moment(time).calendar(null, {
        sameDay: '[Today]',
        nextDay: '[Tomorrow]',
        nextWeek: 'dddd',
        lastDay: '[Yesterday]',
        lastWeek: '[Last] dddd',
        sameElse: 'DD/MM/YYYY'
      });

      ctx.speak(`Sorry, we don't have any availability ${dateString}, can you do another day?`, true);
    } else {
      ctx.speak(`Sorry, we're only open from 9 AM until 5 PM. Is there a different time we can use?`, true);
    }
  }
};


function isWithinBusinessHours(startsAt) {
  let startOfDay = moment(startsAt).startOf("day").add(9, "hours");
  let endOfDay = moment(startOfDay).add(8, "hours");

  return startsAt.isBetween(startOfDay, endOfDay, null, "[)")
}

