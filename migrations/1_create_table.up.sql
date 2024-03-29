CREATE SCHEMA user;

CREATE TABLE user(
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL,
  username VARCHAR(255) NOT NULL,
  avatar VARCHAR(255) NOT NULL,
  activated BOOLEAN DEFAULT FALSE
);

CREATE TABLE address (
  id VARCHAR(255) NOT NULL,
  city VARCHAR(255) NOT NULL,
  address VARCHAR(255) NOT NULL,
  user_id UUID NOT NULL
);
ALTER TABLE address ADD CONSTRAINT address_user_fk FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE ON UPDATE CASCADE;

CREATE TABLE contact (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  phone_number VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  user_id UUID NOT NULL
);
ALTER TABLE address ADD CONSTRAINT contact_user_fk FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE ON UPDATE CASCADE;

CREATE TYPE tp_type AS ENUM('google', 'facebook');
CREATE TABLE tp_association (
  id VARCHAR(255) NOT NULL,
  tp_id VARCHAR(255) NOT NULL,
  tp_type tp_type NOT NULL,
  user_id UUID NOT NULL
);
ALTER TABLE address ADD CONSTRAINT tp_association_user_fk FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE ON UPDATE CASCADE;
