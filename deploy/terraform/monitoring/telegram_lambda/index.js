const https = require('https');

exports.handler = async function (event) {
  const botToken = process.env.TELEGRAM_BOT_TOKEN;
  const chatId = process.env.TELEGRAM_CHAT_ID;

  if (!botToken || !chatId) {
    console.log('Telegram not configured, skipping notification');
    return;
  }

  const snsRecord = event.Records[0].Sns;
  const isAlarm = snsRecord.Subject && snsRecord.Subject.includes('ALARM');
  const isOk = snsRecord.Subject && snsRecord.Subject.includes('OK');

  const icon = isAlarm ? '🚨' : isOk ? '✅' : 'ℹ️';
  const text =
    `${icon} *CoreCheck Alert*\n\n` +
    `*Subject:* ${snsRecord.Subject || 'No subject'}\n\n` +
    `${snsRecord.Message}`;

  await sendTelegramMessage(botToken, chatId, text);
};

function sendTelegramMessage(botToken, chatId, text) {
  const body = JSON.stringify({
    chat_id: chatId,
    text: text,
    parse_mode: 'Markdown',
  });

  return new Promise((resolve, reject) => {
    const req = https.request(
      {
        hostname: 'api.telegram.org',
        path: `/bot${botToken}/sendMessage`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(body),
        },
      },
      (res) => {
        let responseBody = '';
        res.on('data', (chunk) => (responseBody += chunk));
        res.on('end', () => {
          if (res.statusCode === 200) {
            console.log('Telegram message sent successfully');
            resolve();
          } else {
            reject(new Error(`Telegram API returned ${res.statusCode}: ${responseBody}`));
          }
        });
      },
    );

    req.on('error', reject);
    req.write(body);
    req.end();
  });
}
