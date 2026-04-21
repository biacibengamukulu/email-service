
[API Send email]
curl -X POST "https://cloudcalls.easipath.com/backend-email-service/api/v1/send-email" \
-H "Content-Type: application/json" \
-d '{
"from": "no-reply@mails.biacibenga.co.za",
"receiver": [
"sabata@sabata.co.za",
"sabata@joxicraft.co.za"
"biangacila@gmail.com",
"backend@joxicraft.co.za"
],
"subject": "Welcome to our platform - Biacibenga email service",
"html": "<h1>Welcome!</h1><p>Your account has been created successfully.</p>",
"status": "PENDING",
"retries": 0
}'


[API Send Sms Post]
curl -X POST "https://cloudcalls.easipath.com/backend-email-service/api/v1/send-sms/post" \
-H "Content-Type: application/json" \
-d '{"phone":"27684011702","message":"Welcome to our platform - Biacibenga sms service"}'

