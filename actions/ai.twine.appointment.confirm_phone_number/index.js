const numberWordMapping = [
  "oh", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"
];

module.exports["ai.twine.appointment.confirm_phone_number"] = function (ctx, req) {
  const phoneNumber = req.slots["phone-number"];
  if (!phoneNumber) {
    throw new Error("Missing phone number");
  }

  let number = phoneNumber;
  if (phoneNumber.length > 4) {
    let numberWords = phoneNumber.slice(-4).split('').map(digit => {
      const idx = parseInt(digit);
      return numberWordMapping[idx];
    });

    let suffix = numberWords.join(" ");
    ctx.speak(`Is this number ending in ${suffix} a good way to reach you?`, true);
  } else {
    ctx.speak(`Is this a good way to reach you?`, true);
  }
};
