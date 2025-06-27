graph TB
subgraph root["root"]
random_id_deployment_id["random_id.deployment_id<br/>[CREATE]"]
random_password_app_secret["random_password.app_secret<br/>[CREATE]"]
random_string_session_token["random_string.session_token<br/>[CREATE]"]
random_uuid_correlation_id["random_uuid.correlation_id<br/>[CREATE]"]
output_web_instance_id((output.web_instance_id<br/>[NO-OP]))
output_app_secret((output.app_secret<br/>[NO-OP]))
output_correlation_id((output.correlation_id<br/>[NO-OP]))
output_db_name((output.db_name<br/>[NO-OP]))
output_deployment_id((output.deployment_id<br/>[NO-OP]))
output_deployment_tag((output.deployment_tag<br/>[NO-OP]))
output_network_id((output.network_id<br/>[NO-OP]))
output_random_summary((output.random_summary<br/>[NO-OP]))
output_subnet_id((output.subnet_id<br/>[NO-OP]))
output_db_instance_connection_name((output.db_instance_connection_name<br/>[NO-OP]))
output_db_user((output.db_user<br/>[NO-OP]))
output_resource_prefix((output.resource_prefix<br/>[NO-OP]))
output_session_token((output.session_token<br/>[NO-OP]))
output_web_instance_external_ip((output.web_instance_external_ip<br/>[NO-OP]))
var_project[/var.project<br/>[NO-OP]/]
var_region[/var.region<br/>[NO-OP]/]
var_zone[/var.zone<br/>[NO-OP]/]
end

subgraph module_app["module.app"]
module_app_google_compute_firewall_web["module.app.google_compute_firewall.web<br/>[CREATE]"]
module_app_google_compute_instance_web["module.app.google_compute_instance.web<br/>[CREATE]"]
end

subgraph module_network["module.network"]
module_network_google_compute_network_main["module.network.google_compute_network.main<br/>[CREATE]"]
module_network_google_compute_subnetwork_public["module.network.google_compute_subnetwork.public<br/>[CREATE]"]
end

subgraph module_app_module_database["module.app.module.database"]
module_app_module_database_google_sql_database_app["module.app.module.database.google_sql_database.app<br/>[CREATE]"]
module_app_module_database_google_sql_database_instance_main["module.app.module.database.google_sql_database_instance.main<br/>[CREATE]"]
module_app_module_database_google_sql_user_app["module.app.module.database.google_sql_user.app<br/>[CREATE]"]
end

random_password_app_secret --> random_id_deployment_id
random_string_session_token --> random_id_deployment_id
random_uuid_correlation_id --> random_id_deployment_id
module_app_module_database_google_sql_database_instance_main --> module_network_google_compute_network_main
module_app_google_compute_firewall_web --> module_network_google_compute_network_main
module_app_google_compute_instance_web --> module_network_google_compute_network_main
module_app_module_database_google_sql_user_app --> module_network_google_compute_network_main
module_app_module_database_google_sql_database_app --> module_network_google_compute_network_main
module_app_module_database_google_sql_database_app --> module_network_google_compute_subnetwork_public
module_app_module_database_google_sql_database_instance_main --> module_network_google_compute_subnetwork_public
module_app_google_compute_firewall_web --> module_network_google_compute_subnetwork_public
module_app_google_compute_instance_web --> module_network_google_compute_subnetwork_public
module_app_module_database_google_sql_user_app --> module_network_google_compute_subnetwork_public
module_app_module_database_google_sql_database_app --> module_app_module_database_google_sql_database_instance_main
module_app_module_database_google_sql_user_app --> module_app_module_database_google_sql_database_instance_main
module_network_google_compute_subnetwork_public --> module_network_google_compute_network_main
