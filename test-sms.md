
SMS_GATEWAY_URL=https://sms01.umsg.co.za/xml/send
SMS_GATEWAY_USERNAME=kzn_liquor_sa
SMS_GATEWAY_PASSWORD=7dy6tY#D

curl https://sms01.umsg.co.za/xml/send/?number=+27761234567&message=Hello%20from%20your%20Twilio%20powered%20site!
curl -X POST -u $SMS_GATEWAY_USERNAME:$SMS_GATEWAY_PASSWORD $SMS_GATEWAY_URL -d "to=+27761234567&from=+27761234567&text=Hello%20from%20your%20Twilio%20powered%20site!"
