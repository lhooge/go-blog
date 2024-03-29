Go-Blog
====

Small sample open-source blog written in Go using markdown for writing articles and sites.

Prerequisites
--------

 * SQLite3


Configuration
--------
 * create a "custom" folder in the app directory
 * copy go-blog.conf into custom/ folder
 * edit custom/go-blog.conf to your needs
 * create a user for running this application
 * see "examples/systemd" for creating a service 

### SQLite setup ###

Create the initial database (switch to folder clt/): 

~~~
./init_database -sqlite /path/to/your/sqlite/database 
~~~

Change the config file to point to the correct sqlite database

~~~
sqlite_file = /path/to/your/sqlite/database
~~~

### Create user with administration rights ###

Create your first administrator account with createuser (switch to folder clt/):

~~~
./createuser -admin -sqlite /path/to/your/sqlite/database -username test -email test@example.com -displayname "Hello World" 
~~~

Make sure -admin is set. Enter a password for the user.

Licence
-------
    The MIT License (MIT)
    Copyright (c) 2018 Lars Hoogestraat

    Permission is hereby granted, free of charge, to any person obtaining a copy of this software
    and associated documentation files (the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
    and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
    subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all copies or substantial
    portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
    LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
    IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
    SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
