This is a service that simulates generating an access/refresh token pair, when a user is trying to login.

Running the application requires the existence of a redis instance on the working machine.
- &nbsp; To run redis in a docker, from the working directory you can run the command: **make redis**

There is also a configuration file, **config.yaml**, where you can set information about the redis instance, 
but also about the user who will try to log in to the application
- &nbsp; Please be sure that the configuration file contains correct information

If the above steps have been passed successfully, the application can be run using the command: **make all**

To run the tests and see the coverage, please run: **make test**
- &nbsp; Running the tests will delete the existing entries in redis
