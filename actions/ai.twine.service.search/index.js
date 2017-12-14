const mongoose = require("mongoose");

const ServiceSchema = new mongoose.Schema({
  title: String
});

module.exports["ai.twine.service.search"] = function (ctx, req) {
  const db = ctx.database
  const Service = ctx.database.model('Service', ServiceSchema);

  if (!req.slots) {
    return
  }

  const slot = req.slots.fields["service_name"];
  if (!slot || !slot.stringValue) {
    return
  }

  const term = slot.stringValue;
  ctx.speak(`Looking up service name: ${term}`, true)

  return Service.findOne({ $text: { $search: term } }, { score: { $meta: "textScore" } })
    .sort({ score: { $meta: "textScore" } })
    .then((result) => {
      if (result) {
        ctx.speak(`Found service: ${result.title}`, true)
      } else {
        ctx.speak("Could not find service", true)
      }
    })
};
