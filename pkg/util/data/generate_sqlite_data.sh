#!/usr/bin/env bash

# create table reference: https://www.sqlite.org/lang_createtable.html
# ---
# CREATE TABLE [IF NOT EXISTS] [schema_name].table_name (
#	    column_1 data_type PRIMARY KEY,
#   	column_2 data_type NOT NULL,
#	    column_3 data_type DEFAULT 0,
#	    table_constraints
#) [WITHOUT ROWID];

db="sqlite3-data.db"
read -r -d '' create_users << EOF
CREATE TABLE IF NOT EXISTS user (\n
\tid INTEGER PRIMARY KEY,\n
\tfirst_name TEXT NOT NULL,\n
\tlast_name TEXT NOT NULL,\n
\temail TEXT NOT NULL UNIQUE\n
);\n
EOF

read -r -d '' create_address << EOF
CREATE TABLE IF NOT EXISTS address (\n
\tid INTEGER PRIMARY KEY,\n
\tstreet TEXT NOT NULL,\n
\tcity TEXT NOT NULL,\n
\tstate TEXT NOT NULL,\n
\tzip TEXT NOT NULL\n
);\n
EOF

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
EOF

sqlite3 "$db" "$(echo -en $create_users)"
sqlite3 "$db" "$(echo -en $create_address)"
sqlite3 "$db" "$(echo -en $create_user_address)"
echo "listing tables"
echo "--------------"
sqlite3 "$db" .tables

# insert statement
#sqlite3 "$db" "INSERT INTO $table VALUES ('$fname', '$lname', '$email', '$street', '$city', '$state', '$zip')"

