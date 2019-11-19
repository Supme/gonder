SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";

CREATE TABLE `version` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `number` varchar(20) NOT NULL,
  `at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
INSERT INTO `version` (`number`) VALUES ('0.16.3');

CREATE TABLE IF NOT EXISTS `attachment` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `campaign_id` int(11) NOT NULL,
  `path` text NOT NULL,
  PRIMARY KEY (id),
  KEY `campaign_id` (`campaign_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `auth_right` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `auth_unit` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `auth_unit_right` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `auth_unit_id` int(11) NOT NULL,
  `auth_right_id` int(11) NOT NULL,
  PRIMARY KEY (id),
  KEY `auth_unit_id` (`auth_unit_id`),
  KEY `auth_right_id` (`auth_right_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `auth_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `auth_unit_id` int(11) NOT NULL,
  `name` text NOT NULL,
  `password` text NOT NULL COMMENT 'sha256',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `auth_user_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `auth_user_id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  PRIMARY KEY (id),
  KEY `auth_user_id` (`auth_user_id`),
  KEY `group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `campaign` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL,
  `profile_id` int(11) NOT NULL,
  `sender_id` int(11) NOT NULL,
  `name` varchar(300) NOT NULL,
  `subject` varchar(300) NOT NULL,
  `template_html` mediumtext NOT NULL,
  `template_text` mediumtext NOT NULL,
  `template_amp` mediumtext NOT NULL,
  `start_time` timestamp NULL DEFAULT NULL,
  `end_time` timestamp NULL DEFAULT NULL,
  `send_unsubscribe` tinyint(1) NOT NULL DEFAULT '0',
  `accepted` tinyint(1) NOT NULL DEFAULT '0',
  `compress_html` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id),
  KEY `group_id` (`group_id`),
  KEY `from_id` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `group` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `template` VARCHAR(100),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `jumping` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `campaign_id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `url` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY `campaign_id` (`campaign_id`),
  KEY `recipient_id` (`recipient_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `parameter` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `recipient_id` int(11) NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (id),
  KEY `recipient_id` (`recipient_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `recipient` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `campaign_id` int(11) NOT NULL,
  `email` varchar(100) NOT NULL,
  `name` varchar(100) NOT NULL,
  `status` varchar(2000) NULL DEFAULT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_agent` varchar(300),
  `web_agent` varchar(300),
  `removed` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id),
  KEY `campaign_id` (`campaign_id`),
  INDEX `date_status` (`date`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `sender` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL,
  `email` VARCHAR(100) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `utm_url` VARCHAR(100) NOT NULL DEFAULT '',
  `bimi_selector` VARCHAR(20) NOT NULL DEFAULT '',
  `dkim_selector` VARCHAR(20) NOT NULL,
  `dkim_key` VARCHAR(2000) NOT NULL,
  `dkim_use` BOOLEAN NOT NULL,
  PRIMARY KEY (id),
  KEY `group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `unsubscribe` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` VARCHAR(100) NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY `group_id` (`group_id`),
  KEY `campaign_id` (`campaign_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `unsubscribe_extra` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `unsubscribe_id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (id),
  KEY `unsubscribe_id` (`unsubscribe_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `question` (
  `id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `question_data` (
  `id` int(11) NOT NULL,
  `question_id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE `attachment`
  ADD CONSTRAINT `attachment_ibfk_1`
    FOREIGN KEY (`campaign_id`)
    REFERENCES `campaign` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `auth_user_group`
  ADD CONSTRAINT `auth_user_group_ibfk_1`
    FOREIGN KEY (`auth_user_id`)
    REFERENCES `auth_user` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `auth_user_group_ibfk_2`
    FOREIGN KEY (`group_id`)
    REFERENCES `group` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `campaign`
  ADD CONSTRAINT `campaign_ibfk_1`
    FOREIGN KEY (`group_id`)
    REFERENCES `group` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `jumping`
  ADD CONSTRAINT `jumping_ibfk_1`
    FOREIGN KEY (`campaign_id`)
    REFERENCES `campaign` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `jumping_ibfk_2`
    FOREIGN KEY (`recipient_id`)
    REFERENCES `recipient` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `parameter`
  ADD CONSTRAINT `parameter_ibfk_1`
    FOREIGN KEY (`recipient_id`)
    REFERENCES `recipient` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `recipient`
  ADD CONSTRAINT `recipient_ibfk_1`
    FOREIGN KEY (`campaign_id`)
    REFERENCES `campaign` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `sender`
  ADD CONSTRAINT `sender_ibfk_1`
    FOREIGN KEY (`group_id`)
    REFERENCES `group` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `unsubscribe`
  ADD CONSTRAINT `unsubscribe_ibfk_1`
    FOREIGN KEY (`group_id`)
    REFERENCES `group` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `unsubscribe_extra`
  ADD CONSTRAINT `unsubscribe_extra_ibfk_1`
    FOREIGN KEY (`unsubscribe_id`)
    REFERENCES `unsubscribe` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `question`
  ADD PRIMARY KEY (`id`),
  ADD KEY `recipient_id` (`recipient_id`);
ALTER TABLE `question`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `question`
  ADD CONSTRAINT `question_ibfk_1`
    FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `question_data`
  ADD PRIMARY KEY (`id`), ADD KEY `question_id` (`question_id`);
ALTER TABLE `question_data`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `question_data`
  ADD CONSTRAINT `question_data_ibfk_1`
    FOREIGN KEY (`question_id`) REFERENCES `question` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

INSERT INTO `auth_right` (`id`, `name`) VALUES
  (1, 'get-groups'),
  (2, 'save-groups'),
  (3, 'add-groups'),
  (4, 'get-campaigns'),
  (5, 'save-campaigns'),
  (6, 'add-campaigns'),
  (7, 'get-campaign'),
  (8, 'save-campaign'),
  (9, 'get-recipients'),
  (10, 'get-recipient-parameters'),
  (11, 'upload-recipients'),
  (12, 'delete-recipients'),
  (13, 'get-profiles'),
  (14, 'add-profiles'),
  (15, 'delete-profiles'),
  (16, 'save-profiles'),
  (17, 'accept-campaign'),
  (18, 'get-log-main'),
  (19, 'get-log-api'),
  (20, 'get-log-campaign'),
  (21, 'get-log-utm');

INSERT INTO `auth_unit` (`id`, `name`) VALUES
  (0, 'administrator'),
  (1, 'accepter');

INSERT INTO `auth_unit_right` (`id`, `auth_unit_id`, `auth_right_id`) VALUES
  (1, 1, 1),
  (2, 1, 4),
  (3, 1, 7),
  (4, 1, 9),
  (5, 1, 10),
  (6, 1, 17);

INSERT INTO `auth_user` (`id`, `auth_unit_id`, `name`, `password`) VALUES
  (1, 0, 'admin', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918'),
  (2, 1, 'user', '04f8996da763b7a969b1028ee3007569eaf3a635486ddab211d512c85b9df8fb');
