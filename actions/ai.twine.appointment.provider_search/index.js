const twine = require("twine-action-lib-service");
const ngram = require("n-gram");

module.exports["ai.twine.appointment.provider_search"] = function (ctx, req) {
  twine.registerModels(ctx);

  const db = ctx.database

  console.log(req);

  let term = null;
  if (req.slots && req.slots["provider_name"]) {
    term = req.slots["provider_name"];
  }

  if (!term) {
    return;
  }

  ctx.setSlot("found_provider_name", term);
};
