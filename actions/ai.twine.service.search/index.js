const twine = require("twine-action-lib-service");

module.exports["ai.twine.service.search"] = function (ctx, req) {
  twine.registerModels(ctx);

  const db = ctx.database

  if (!req.slots) {
    return;
  }

  const term = req.slots["service_name"];
  if (!term) {
    return;
  }

  ctx.logger.debug(`Looking up service name: ${term}`)

  return ctx.models.Service.findOne(
      { $text: { $search: term } },
      { score: { $meta: "textScore" } }
    )
    .sort({ score: { $meta: "textScore" } })
    .then((result) => {
      if (result.toObject()) {
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
    })
};
