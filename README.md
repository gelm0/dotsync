# dotsync
Keep your files synced
## TODO
We need to finish backend this means full functionality. Most likely we are implementing it as a backend API server with a Electron frontend.


## Architecture
### Backend endpoints
| Operation | Endpoint            | Returns                                 | Description                    |
|-----------|---------------------|-----------------------------------------|--------------------------------|
| POST      | /v1/config          | Id of new config                        | Creates a new syncconfig       |
| PUT       | /v1/config{id}      | Void                                    | Updates a current              |
| GET       | /v1/config/{id}     | Syncconfig                              | Gets a synconfig json object   |
| GET       | /v1/config          | All Syncconfigs                         | Gets all synconfig json object |
| DELETE    | /v1/config/{id}     | Void                                    | Deletes a syncconfig           |
| GET       | /v1/file/{name}     | IDs of config related to filename       | Gets a synconfig json object   |
| POST      | /v1/sync/{id}       | Current file as stream being synced     | Syncs a syncconfig             |
| POST      | /v1/sync            | Current file as stream being synced     | Syncs all syncconfigs          |

```
|=================|                         |==================|
|                 |                         |                  |
| dotsync backend |=>TCP/11000------------0=| dotsync frontend |
|                 |                         |                  |
|=================|                         |==================|
        |                                             |
        |                                |============================|                       
        |                                | Frontend config            |
        |                                | stored in ~/.dotsync/config| 
        |                                | Keeps track of current     |
        |                                | Watched files              |
        |                                |                            |                   
        |                                |============================|
    |===========================|                       
    | storage config            |
    | stored in ~/.dotsync/state| 
    |                           |                   
    |                           |
    |===========================|
```
