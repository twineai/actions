const twine = require("twine-action-lib-service");

module.exports["ai.twine.core.get_current_phone_number"] = function (ctx, req) {
  twine.registerModels(ctx);

  const cid = req.conversationId;
  if (!cid) {
    throw new Error("Missing conversation ID");
  }

  return ctx.models.Conversation
    .findOne({
      _id: cid,
    })
    .then((conversation) => {
      ctx.logger.debug(`Found conversation: ${conversation}`);

      if (!conversation) { return null; }
      if (!conversation.meta) { return null; }
      if (!conversation.meta.source) { return null; }

      return conversation.meta.source.tel;
    })
    .then((phoneNumber) => {
      if (phoneNumber) {
        ctx.setSlot("phone-number", phoneNumber);
      } else {
        ctx.setSlot("phone-number", null);
      }
    });
};
