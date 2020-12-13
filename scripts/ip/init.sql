DROP TABLE IF EXISTS `ip_details`;
CREATE TABLE `ip_details` (
  `uuid` varchar(255) DEFAULT NULL,
  `created_at` varchar(255) DEFAULT NULL,
  `updated_at` varchar(255) DEFAULT NULL,
  `ip_address` varchar(255) DEFAULT NULL,
  `response_code` varchar(255) DEFAULT NULL,
  UNIQUE KEY `ip_address` (`ip_address`)
) 
