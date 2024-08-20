There are 2 users in my design: User(student) and admin.

User can only update, delete and create only their own values, and even though they have token, they cannot manipulate other students data.

Admin can delete, create, update and delete any user using his own token.

So when token is being passed, that is resolved to get to know the user_id, using which userType will be determined. I am passing this userType from transport layer to Db layer(instead of user_id as suggested), so when create or update is done, userType will let the db know the values of created_by and updated_by instead of explicity mentioning it duirng creation or updation.

I am loading the admin details and database details from the env file.

I am logging all the operations performed and errors if any in the app.log file.

The admin token is predefined in the env file. So there is only one admin, whoes token is fixed, so when this token is encountered all the operations are allowed.

The token for user is generated after registering, which is unique for every user. This token contains the details of the user_id, which is resolved in the authorization middleware and if the user_id extracted from the token and the user_id on which any operation is requested are the same, then it is allowed else, authorization is denied, since this user cannot access other users details or manipulate them.

User_id in my db is auto-incremented, hence it is not being passed through context from transport layer to db layer.

To be able to get the user_id during token generation from the db, "GetLastInsertedStudentID()" function is used.

admin token is : eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE2ODAxMTM3OTF9.QoNz4Q2KZdGRPQQ1FQOmD2ZWmr_9hwnj5FkZQWXLnxA

For any get request only the token authentication is done, it doesn't check for the userType because every student(user) should be able to view others details but should not be allowed to modify others details.

Service layer, is used to establish contact between the transport layer and the database layer.


Sample curl commands :
POST request :
C:\Users\ADMIN>curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d "{\"name\": \"Surya\", \"course\": \"Computer Science\", \"grade\": \"A\"}"
A token will be returned after registering.
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMiwibmFtZSI6IlN1cnlhIiwiZXhwIjoxNzI0MzkyNTk0LCJpYXQiOjE3MjQxMzMzOTR9.J8b_Yt1UTwgixZDpScQKHb1UD-UZ6aG03nRqhYK8DVs"}
The above token can be used only by this specific user to perform deletion or updation of his own user_id and cannot manipulate other data.


PUT request:
C:\Users\ADMIN>curl -X PUT http://localhost:8080/students/11 -H "Content-Type: application/json" -H "Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMSwibmFtZSI6IlN1cnlhIiwiZXhwIjoxNzI0MzEwMTI2LCJpYXQiOjE3MjQwNTA5MjZ9.6KFZhYCkJlTScC97I15FKR_JvXkA2ymyQ_sHtHplE2I" -d "{\"name\": \"Surya Dev\", \"course\": \"Updated Data Science\", \"grade\": \"A+\"}"
{"id":11,"password":"","name":"Surya Dev","course":"Updated Data Science","grade":"A+","created_by":"","created_on":"0001-01-01T00:00:00Z","updated_by":"user","updated_on":"0001-01-01T00:00:00Z"}


DELETE request:
C:\Users\ADMIN>curl -X DELETE http://localhost:8080/students/12 -H "Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE2ODAxMTM3OTF9.QoNz4Q2KZdGRPQQ1FQOmD2ZWmr_9hwnj5FkZQWXLnxA"

GET request:
C:\Users\ADMIN>curl -X GET http://localhost:8080/students/7 -H "Authorization: Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNCwibmFtZSI6IlNhbmFudGgiLCJleHAiOjE3MjQzOTM1NTAsImlhdCI6MTcyNDEzNDM1MH0.WPxCJQK2TVOTvvwvRcs-3QPZUcxiBZ_zCVa53zDBMh4"
{"id":7,"password":"","name":"Rahul I V","course":"Data Science","grade":"A+","created_by":"user","created_on":"2024-08-17T19:38:10Z","updated_by":"admin","updated_on":"2024-08-18T22:14:34Z"}

So any valid token can be used to get the students details. Only the authentication of token is done, but no userType is deduced to restrict any get requests.

This is the DB Schema used to store student details:
user_id	int	NO	PRI		auto_increment
password	varchar(150)	YES			
name	varchar(100)	NO			
course	varchar(100)	YES			
grade	varchar(10)	YES			
created_by	varchar(50)	YES			
created_on	timestamp	YES		CURRENT_TIMESTAMP	DEFAULT_GENERATED
updated_by	varchar(50)	YES			
updated_on	timestamp	YES		CURRENT_TIMESTAMP	DEFAULT_GENERATED on update CURRENT_TIMESTAMP
