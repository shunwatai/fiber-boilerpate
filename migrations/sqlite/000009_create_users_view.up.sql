DROP VIEW IF EXISTS users_view;
CREATE VIEW users_view AS 
SELECT 
  id,
  name,
  password,
  email,
  first_name,
  last_name,
  disabled,
  is_oauth,
  provider,
  (coalesce(name,'') || ' ' || coalesce(first_name,'') || ' ' || coalesce(last_name,'') || ' ' || coalesce(email,'')) as search,
  created_at,
  updated_at
from users;

