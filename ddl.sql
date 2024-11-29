CREATE DATABASE API_GC1;

CREATE TABLE IF NOT EXISTS Customers (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  email VARCHAR(50) UNIQUE NOT NULL,
  phone VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON Customers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 

insert into
Customers (name, email, phone)
values
  (
    'Kazuya Mishima',   
    'kazuya.mishima@gmail.com',
    '+6212345678'
  ),
  (
    'Troy Baker',    
    'troy.baker@example.com',
    '+6223458901'
  ),
  (
    'Dominus Vexen',    
    'dominus.vexen@example.com',
    '+6256789012'
  );