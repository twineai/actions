const twine = require("twine-action-lib-service");
const ngram = require("n-gram");

module.exports["ai.twine.service.lookup"] = function (ctx, req) {
  twine.registerModels(ctx);

  const db = ctx.database

  if (!req.slots) {
    return;
  }

  const term = req.slots["service_name"];
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
      if (result) {
        console.log(result);
        ctx.logger.debug(`Found service: ${result}`);
        ctx.setSlot("found_service_data", {
          module: "ai.twine.service",
          businessId: twine.TempConstants.businessId,
          id: result.id,
        });
      } else {
        ctx.logger.debug("Could not find service")
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
