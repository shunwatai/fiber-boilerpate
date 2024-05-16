DROP VIEW IF EXISTS group_resource_acls_view;
CREATE OR REPLACE VIEW group_resource_acls_view AS 
SELECT 
  group_resource_acls.id id,
  group_id,
  groups.name group_name,
  resource_id,
  resources.name resource_name,
  permission_type_id,
  permission_types.name permission_type,
  group_resource_acls.created_at created_at,
  group_resource_acls.updated_at updated_at
from group_resource_acls
left join groups on group_id = groups.id
left join resources on resource_id = resources.id
left join permission_types on permission_type_id = permission_types.id;
