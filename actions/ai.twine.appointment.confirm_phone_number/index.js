
module.exports["ai.twine.appointment.confirm_phone_number"] = function (ctx, req) {
  const phoneNumber = req.slots["phone-number"];
  if (!phoneNumber) {
    throw new Error("Missing phone number");
  }

  let number = phoneNumber;
  if (phoneNumber.length > 4) {
    number = phoneNumber.slice(-4);
  }

  ctx.speak(`Is this number ending in ${number} a good way to reach you?`, true);
};
