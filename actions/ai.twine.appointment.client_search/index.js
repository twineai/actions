const twine = require("twine-action-lib-service");
const ngram = require("n-gram");

module.exports["ai.twine.appointment.client_search"] = function (ctx, req) {
  twine.registerModels(ctx);

  const db = ctx.database

  console.log(req);

  let term = req.text;
  if (req.slots) {
    if (req.slots["client_name"]) {
      term = req.slots["client_name"];
    } else if (req.slots["person_name"]) {
      term = req.slots["person_name"];
    }
  }

  if (term) {
    ctx.setSlot("found_client_name", term);
  } else {
    ctx.setSlot("found_client_name", null);
  }
};
