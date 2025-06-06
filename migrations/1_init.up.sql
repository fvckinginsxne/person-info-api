CREATE TABLE IF NOT EXISTS people (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    surname VARCHAR(255) NOT NULL,
    patronymic VARCHAR(255),
    age INT NOT NULL,
    gender VARCHAR(16) NOT NULL,
    nationality VARCHAR(32) NOT NULL
);

CREATE INDEX IF NOT EXISTS people_name_idx ON people (name);
CREATE INDEX IF NOT EXISTS people_surname_idx ON people (surname);
CREATE INDEX IF NOT EXISTS people_patronymic_idx ON people (patronymic);