#!/usr/bin/env bash

# database file name
db="sqlite3-data.db"

# create users table and indexes
read -r -d '' create_users << EOF
CREATE TABLE IF NOT EXISTS user (\n
\tid INTEGER PRIMARY KEY,\n
\tfirst_name TEXT NOT NULL,\n
\tlast_name TEXT NOT NULL,\n
\temail TEXT NOT NULL UNIQUE\n
);\n
CREATE INDEX IF NOT EXISTS 'idx_user_pks' ON 'user' (id);
EOF

# create address table and indexes
read -r -d '' create_address << EOF
CREATE TABLE IF NOT EXISTS address (\n
\tid INTEGER PRIMARY KEY,\n
\tstreet TEXT NOT NULL,\n
\tcity TEXT NOT NULL,\n
\tstate TEXT NOT NULL,\n
\tzip TEXT NOT NULL\n
);\n
CREATE INDEX IF NOT EXISTS 'idx_address_pks' ON 'address' (id);
EOF

# create table and indexes
read -r -d '' create_user_address << EOF
CREATE TABLE IF NOT EXISTS user_address (\n
  \tuser_id INTEGER,\n
  \taddress_id INTEGER,\n
  \tPRIMARY KEY(user_id, address_id),\n
  \tFOREIGN KEY(user_id)\n
     \t\tREFERENCES users (id)\n
      \t\t\tON DELETE CASCADE\n
      \t\t\tON UPDATE NO ACTION,\n
  \tFOREIGN KEY(address_id)\n
    \t\tREFERENCES address (id)\n
      \t\t\tON DELETE CASCADE\n
      \t\t\tON UPDATE NO ACTION\n
);\n
CREATE INDEX IF NOT EXISTS 'idx_user_address_pks' ON 'user_address' (user_id, address_id);
EOF

# actually run the commands to create the tables and indexes
sqlite3 "$db" "$(echo -en $create_users)"
sqlite3 "$db" "$(echo -en $create_address)"
sqlite3 "$db" "$(echo -en $create_user_address)"

# import user data into the user table
sqlite3 "$db" ".mode csv" ".import user_data.csv user" ".exit"

# import address data into the address table
sqlite3 "$db" ".mode csv" ".import address_data.csv address" ".exit"

# drop the database
rm "$db"

