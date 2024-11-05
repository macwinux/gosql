# GOSQL 

## GoSQL is a demo database built in go for learning purposes 

I'm following Phil Eaton repository but changing some things because I found some unnecessary pointers and trying to 
simplify it.

# Example

```bash
$ cd gosql
$ go run cmd/main.go
Welcome to gosql.
# CREATE TABLE users (id INT, name TEXT, age INT);
ok
# INSERT INTO users VALUES (1, "Carlos", 33);
ok
# SELECT id, name, age FROM users;
  id | name     | age
-----+----------+------
   1 | Carlos   |  33
ok
```

