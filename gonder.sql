SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

CREATE TABLE IF NOT EXISTS `attachment` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `path` text NOT NULL,
  `file` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_right` (
  `id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_unit` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_unit_right` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `auth_right_id` int(11) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_user` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `name` text NOT NULL,
  `password` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_user_group` (
  `id` int(11) NOT NULL,
  `auth_user_id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `campaign` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `profile_id` int(11) NOT NULL,
  `from` text NOT NULL,
  `from_name` text NOT NULL,
  `name` text NOT NULL,
  `subject` text NOT NULL,
  `body` text NOT NULL,
  `start_time` timestamp NULL DEFAULT NULL,
  `end_time` timestamp NULL DEFAULT NULL,
  `send_unsubscribe` varchar(1) NOT NULL DEFAULT 'n'
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `group` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL,
  `template` text
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `jumping` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `url` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `parameter` (
  `id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=206285 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `profile` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL,
  `iface` text NOT NULL COMMENT 'Example: xx.xx.xx.xx or socks://ip:port for socks5. Blank for default interface.',
  `host` text NOT NULL COMMENT 'The name of the server on behalf of which there is a sending.',
  `stream` int(11) NOT NULL COMMENT 'Number of concurrent streams',
  `resend_delay` int(11) NOT NULL,
  `resend_count` int(11) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `recipient` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  `status` text,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_agent` text,
  `web_agent` text,
  `removed` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB AUTO_INCREMENT=168848 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `status` (
  `id` int(3) NOT NULL,
  `pattern` varchar(250) NOT NULL,
  `bounce_id` int(1) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=264 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `unsubscribe` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8;


ALTER TABLE `attachment`
ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `auth_right`
ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `id` (`id`), ADD KEY `id_2` (`id`);

ALTER TABLE `auth_unit`
ADD PRIMARY KEY (`id`);

ALTER TABLE `auth_unit_right`
ADD PRIMARY KEY (`id`);

ALTER TABLE `auth_user`
ADD PRIMARY KEY (`id`);

ALTER TABLE `auth_user_group`
ADD PRIMARY KEY (`id`);

ALTER TABLE `campaign`
ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`), ADD KEY `profile_id` (`profile_id`);

ALTER TABLE `group`
ADD PRIMARY KEY (`id`);

ALTER TABLE `jumping`
ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`), ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `parameter`
ADD PRIMARY KEY (`id`), ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `profile`
ADD PRIMARY KEY (`id`), ADD KEY `id` (`id`);

ALTER TABLE `recipient`
ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `status`
ADD PRIMARY KEY (`id`);

ALTER TABLE `unsubscribe`
ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`), ADD KEY `campaign_id` (`campaign_id`);


ALTER TABLE `attachment`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=2;
ALTER TABLE `auth_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=17;
ALTER TABLE `auth_unit`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=2;
ALTER TABLE `auth_unit_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=9;
ALTER TABLE `auth_user`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=3;
ALTER TABLE `auth_user_group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=2;
ALTER TABLE `campaign`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=8;
ALTER TABLE `group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=21;
ALTER TABLE `jumping`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `parameter`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=206285;
ALTER TABLE `profile`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=16;
ALTER TABLE `recipient`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=168848;
ALTER TABLE `status`
MODIFY `id` int(3) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=264;
ALTER TABLE `unsubscribe`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=15;

ALTER TABLE `attachment`
ADD CONSTRAINT `attachment_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `campaign`
ADD CONSTRAINT `campaign_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `campaign_ibfk_2` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `campaign_ibfk_3` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `jumping`
ADD CONSTRAINT `jumping_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `jumping_ibfk_2` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `parameter`
ADD CONSTRAINT `parameter_ibfk_1` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `recipient`
ADD CONSTRAINT `recipient_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `unsubscribe`
ADD CONSTRAINT `unsubscribe_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;



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
  (16, 'save-profiles');

INSERT INTO `auth_unit` (`id`, `name`) VALUES
  (1, 'reader');

INSERT INTO `auth_user` (`id`, `auth_unit_id`, `name`, `password`) VALUES
  (1, 0, 'admin', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918'),
  (2, 1, 'user', '04f8996da763b7a969b1028ee3007569eaf3a635486ddab211d512c85b9df8fb');

INSERT INTO `auth_unit_right` (`id`, `auth_unit_id`, `auth_right_id`) VALUES
  (1, 1, 1),
  (2, 1, 4),
  (3, 1, 7),
  (4, 1, 9),
  (5, 1, 10),
  (6, 1, 11),
  (7, 1, 12),
  (8, 1, 8);