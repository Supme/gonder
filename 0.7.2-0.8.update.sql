SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

ALTER TABLE `recipient`
  CHANGE `status` `status` VARCHAR(2000) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  CHANGE `email` `email` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  CHANGE `name` `name` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  CHANGE `client_agent` `client_agent` VARCHAR(300) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  CHANGE `web_agent` `web_agent` VARCHAR(300) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL;
ALTER TABLE `recipient` ADD INDEX `date_status` (`date`, `status`);

ALTER TABLE `campaign`
  CHANGE `name` `name` VARCHAR(300) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  CHANGE `subject` `subject` VARCHAR(300) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL;

ALTER TABLE `group`
  CHANGE `name` `name` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  CHANGE `template` `template` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL;

ALTER TABLE `sender`
  CHANGE `email` `email` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  CHANGE `name` `name` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL;

ALTER TABLE `unsubscribe`
  CHANGE `email` `email` VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL;

