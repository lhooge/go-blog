Go-Blog
====

Small sample open-source blog written in Go using markdown for formatting articles and additional sites.

* Articles can be created by users or administrators
  * Administrator can manage all articles; users can only manage personal articles

* Files can be uploaded by users or administrators
  * Administrator can manage all files; users can only manage personal files

* Users can be created (by administrator)

* Additional sites can be created (by administrator)

Not really for production use; started for learning Go for web development.


Prerequisites
--------

 * SQLite3


Configuration
--------

 * copy go-blog.conf into custom/ folder
 * edit custom/go-blog.conf to your needs

### SQLite setup ###

The configuration for sqlite is simple:

~~~
  database_engine = sqlite
  sqlite_file = /path/to/your/sqlite/database
~~~


### Create user with administration rights ###

Create your first administrator account with create_user:
~~~
./createuser -admin -config {{BLOG_CONFIG}} -username test -email test@example.com -displayname "Hello World" -password secret1234
~~~

Make sure -admin is set.

TODOs
-----
 * Activation link when registering new users
 * Add and fix test
 * Comment user interceptor
 * Revisit Makefile
 * Database update tasks / simple query builder
 * Order possibilities in admin panel
 * Review preview of articles and sites / error handling
 * Support some environmental variables / flags

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
