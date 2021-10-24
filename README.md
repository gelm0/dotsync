# dotsync
Keep your files synced
## TODO
We need to finish backend this means full functionality. Most likely we are implementing it as a backend API server with a Electron frontend.


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