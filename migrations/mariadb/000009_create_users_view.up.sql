DROP VIEW IF EXISTS users_view;
CREATE OR REPLACE VIEW users_view 
AS select id,
name,
password,
email,
first_name,
last_name,
disabled,
is_oauth,
provider,
(CONCAT(IFNULL(name,''),' ',IFNULL(first_name,''),' ',IFNULL(last_name,''),' ',IFNULL(email,''))) as search,
created_at,
updated_at
from users;
