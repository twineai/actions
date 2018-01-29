const twine = require("twine-action-lib-service");
const ngram = require("n-gram");

module.exports["ai.twine.service.search"] = function (ctx, req) {
  twine.registerModels(ctx);

  const db = ctx.database

  console.log(req);

  ctx.logger.debug(`Doing service search: ${req}`)

  let term = req.text;
  if (req.slots && req.slots["service_name"]) {
    term = req.slots["service_name"];
  }

  if (!term) {
    return;
  }

  ctx.logger.debug(`Looking up service name: ${term} - ${ngram.trigram(term)}`);

  return Promise.resolve()
    .then(() => {
      return search(ctx, term);
    })
    .then((result) => {
      if (!result) {
        return search(ctx, ngram.trigram(term).join(" "));
      } else {
        return result;
      }
    })
    .then((result) => {
      if (!result) {
        ctx.logger.debug("Could not find service")
        return;
      }

      ctx.logger.debug(`Found potential service [${result.id}]: ${result.title} = ${result.get('score')}`);

      if (result.get('score') >= 7) {
        ctx.setSlot("found_service_data", {
          businessId: twine.TempConstants.businessId,
          _id: result.id,
        });
      } else {
        ctx.logger.debug("Score too low, skipping");
      }
    });
};

function search(ctx, term) {
  return ctx.models.Service.findOne(
      { $text: { $search: term } },
      { score: { $meta: "textScore" } }
    )
    .sort({ score: { $meta: "textScore" } });
}
