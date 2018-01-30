const twine = require("twine-action-lib-service");

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

  const businessId = twine.TempConstants.businessId;

  ctx.logger.debug(`Looking up service name: '${term}'`);

  return Promise.resolve()
    .then(() => {
      return search(ctx, businessId, term);
    })
    .then((result) => {
      if (!result) {
        ctx.logger.debug("Could not find service")
        return;
      }

      ctx.logger.debug(`Found service [${result.id}]: ${result.title}`);
      ctx.setSlot("found_service_data", {
        businessId: businessId,
        _id: result.id,
      });
    });
};

function search(ctx, businessId, term) {
  return ctx.elasticsearchClient.search({
    index: "services",
    type: "doc",
    body: {
      min_score: 0.5,
      query: {
        bool: {
          should: [
            {
              match: {
                "title.phonetic": {
                  query: term,
                  fuzziness: "AUTO"
                }
              }
            }
          ],
          filter: {
            term: {
              businessId: businessId
            }
          }
        }
      }
    }
  })
  .then((resp) => {
    const results = resp.hits.hits;
    if (results.length > 0) {
      ctx.logger.debug("Service search results: %j", results);
      return results[0];
    }
    return null;
  })
  .then((result) => {
    if (result) {
      return getService(ctx, result._source.businessId, result._id);
    }
    return null;
  });
}

function getService(ctx, businessId, id) {
  return ctx.models.Service
    .findOne({
      businessId: businessId,
      _id: id,
    })
    .exec();
}