## Available http requests
    POST   /api/v1/user                  - create user by login

    GET    /api/v1/session               - check cookie
    POST   /api/v1/session               - log in by login and password
    
    GET    /api/v1/profile/{user_id}     - get text and photos of profile
    POST   /api/v1/profile/{user_id}     - create profile

    GET    /api/v1/indexed               - get susits pair

    GET    /api/v1/likes/my              - get all likes by user
    GET    /api/v1/likes/me              - get all likes to user
    
    GET    /api/v1/matches               - get all matches
    
    GET    /api/v1/matches/actual        - get new matches
    DELETE /api/v1/matches/actual        - make match viewed

    POST   /api/v1/like/{user_id}        - make like

## Details of http requests
### /api/v1/user
#### POST
    request:
      Cookie: not required
      body: {
        "login":login,
        "password":password,
        "phone_number":number,
        "email":email,
      }
    response:  
      201 - success create user
      400 - bad request
      403 - permission denied, user already exist
      500 - something went wrong :(
### /api/v1/session
#### GET
    request:
      Cookie: token=your_access_token
    response:
      200 - when you send cookie with exist token
      401 - token not found
#### POST
    request:
      Cookie: not required
      body: {
        "login":login,
        "password":password
      }
    response:  
      201 - success logged in
        Set-Cookies: token=your_access_token
      400 - bad request
      403 - permission denied, wrong password
      500 - something went wrong :(
### /api/v1/profile/{user_id}
#### GET
    request:
      Cookie: not required
    response:  
      200 - get info about profile
      body: {
        "profile_text": string,
        "sex": bool (false - woman, true - man),
        "birthday": string,
        "name": string,
        "photo": [string, ...]
      }
      500 - something went wrong
#### POST
    request:
      Cookie: token=your_access_token
      body: {
        "profile_text": string,
        "sex": bool (false - woman, true - man),
        "birthday": string,
        "name": string,
        "photo": [string, ...]
      }
    response:
      201 - sussess created profile
      400 - bad request
      403 - permission denied
      500 - something went wrong
### /api/v1/indexed
#### GET
    request:
      Cookie: token=your_access_token
    response: 
      200 - get user_id
      body: {
        "user_id": int
      }
      400 - bad request
      401 - unauthorized
      403 - 0 likes left for today
      500 - something went wrong
### /api/v1/likes/my
#### GET
    request: 
      Cookie: token=your_access_token
    response:
      200 - get likes
      body: [
        {
          "user_id": int,
          "time": string
        }, ...
      ]
      401 - pUnauthorized
      500 - something went wrong
### /api/v1/likes/me
#### GET
    request: 
      Cookie: token=your_access_token
    response:
      200 - get likes
      body: [
        {
          "user_id": int,
          "time": string
        }, ...
      ]
      401 - Unauthorized
      500 - something went wrong
### /api/v1/matches
#### GET
    request: 
      Cookie: token=your_access_token
    response:
      200 - get matches
      body: [
        {
          "user_id": int,
          "time": string
        }, ...
      ]
      401 - Unauthorized
      500 - something went wrong
### /api/v1/match/actual
#### GET
    request: 
      Cookie: token=your_access_token
    response:
      200 - get matches
      body: [
        {
          "user_id": int,
          "time": string
        }, ...
      ]
      401 - Unauthorized
      403 - 0 left likes for today
      500 - something went wrong
#### DELETE
    request:
      Cookie: token=your_access_token
      body: {
        "user_id": int (with who match)
      }
    response:
      201 - successfully make match viewed
      500 - something went wrong
### /api/v1/like/{user_id}
#### POST
    request:
      Cookie: token=your_access_token
    response:
      201 - successfully liked
      403 - permission denied
      500 - something went wrong