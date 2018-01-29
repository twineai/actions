module.exports["ai.twine.core.accept_phone_number"] = function (ctx, req) {
  if (req.slots) {
    const phoneNumber = req.slots["phone-number"];
    if (phoneNumber) {
      ctx.setSlot("found_phone_number", phoneNumber);
      return;
    }
  }

  throw new Error("Unable to find phone number to accept");
};
