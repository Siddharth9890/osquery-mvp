CREATE TABLE IF NOT EXISTS system_info (
    id INT AUTO_INCREMENT PRIMARY KEY,
    os_version VARCHAR(255) NOT NULL,
    os_name VARCHAR(255) NOT NULL,
    os_platform VARCHAR(255) NOT NULL,
    osquery_version VARCHAR(255) NOT NULL,
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS installed_apps (
    id INT AUTO_INCREMENT PRIMARY KEY,
    system_info_id INT,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(255),
    FOREIGN KEY (system_info_id) REFERENCES system_info(id) ON DELETE CASCADE
);


CREATE INDEX idx_system_info_collected_at ON system_info(collected_at);
CREATE INDEX idx_installed_apps_system_info_id ON installed_apps(system_info_id);