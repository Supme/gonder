SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

CREATE TABLE `attachment` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `path` text NOT NULL,
  `file` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `auth_right` (
  `id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `auth_unit` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `auth_unit_right` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `auth_right_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `auth_user` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `name` text NOT NULL,
  `password` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `auth_user_group` (
  `id` int(11) NOT NULL,
  `auth_user_id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `campaign` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `profile_id` int(11) NOT NULL,
  `sender_id` int(11) NOT NULL DEFAULT '0',
  `name` text NOT NULL,
  `subject` text NOT NULL,
  `body` text NOT NULL,
  `start_time` timestamp NULL DEFAULT NULL,
  `end_time` timestamp NULL DEFAULT NULL,
  `send_unsubscribe` tinyint(1) NOT NULL DEFAULT '0',
  `accepted` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `group` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL,
  `template` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `jumping` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `url` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `parameter` (
  `id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `profile` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL,
  `iface` text NOT NULL COMMENT 'Example: xx.xx.xx.xx or socks://ip:port for socks5. Blank for default interface.',
  `host` text NOT NULL COMMENT 'The name of the server on behalf of which there is a sending.',
  `stream` int(11) NOT NULL COMMENT 'Number of concurrent streams',
  `resend_delay` int(11) NOT NULL,
  `resend_count` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `recipient` (
  `id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  `status` text,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_agent` text,
  `web_agent` text,
  `removed` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `sender` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `status` (
  `id` int(3) NOT NULL,
  `pattern` varchar(250) NOT NULL,
  `bounce_id` int(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `unsubscribe` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `attachment`
ADD PRIMARY KEY (`id`),
ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `auth_right`
ADD PRIMARY KEY (`id`),
ADD UNIQUE KEY `id` (`id`),
ADD KEY `id_2` (`id`);

ALTER TABLE `auth_unit`
ADD PRIMARY KEY (`id`);

ALTER TABLE `auth_unit_right`
ADD PRIMARY KEY (`id`),
ADD KEY `auth_unit_id` (`auth_unit_id`),
ADD KEY `auth_right_id` (`auth_right_id`);

ALTER TABLE `auth_user`
ADD PRIMARY KEY (`id`);

ALTER TABLE `auth_user_group`
ADD PRIMARY KEY (`id`),
ADD KEY `auth_user_id` (`auth_user_id`),
ADD KEY `group_id` (`group_id`);

ALTER TABLE `campaign`
ADD PRIMARY KEY (`id`),
ADD KEY `group_id` (`group_id`),
ADD KEY `profile_id` (`profile_id`);

ALTER TABLE `group`
ADD PRIMARY KEY (`id`);

ALTER TABLE `jumping`
ADD PRIMARY KEY (`id`),
ADD KEY `campaign_id` (`campaign_id`),
ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `parameter`
ADD PRIMARY KEY (`id`),
ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `profile`
ADD PRIMARY KEY (`id`);

ALTER TABLE `recipient`
ADD PRIMARY KEY (`id`),
ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `status`
ADD PRIMARY KEY (`id`);

ALTER TABLE `unsubscribe`
ADD PRIMARY KEY (`id`),
ADD KEY `group_id` (`group_id`),
ADD KEY `campaign_id` (`campaign_id`);


ALTER TABLE `attachment`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `auth_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `auth_unit`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `auth_unit_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `auth_user`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `auth_user_group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `campaign`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `jumping`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `parameter`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `profile`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `recipient`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `status`
MODIFY `id` int(3) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;
ALTER TABLE `unsubscribe`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;

ALTER TABLE `attachment`
ADD CONSTRAINT `attachment_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `auth_unit_right`
ADD CONSTRAINT `auth_unit_right_ibfk_1` FOREIGN KEY (`auth_unit_id`) REFERENCES `auth_unit` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `auth_unit_right_ibfk_2` FOREIGN KEY (`auth_right_id`) REFERENCES `auth_right` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `auth_user_group`
ADD CONSTRAINT `auth_user_group_ibfk_1` FOREIGN KEY (`auth_user_id`) REFERENCES `auth_user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `auth_user_group_ibfk_2` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `campaign`
ADD CONSTRAINT `campaign_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD KEY `sender_id` (`sender_id`);

ALTER TABLE `jumping`
ADD CONSTRAINT `jumping_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `jumping_ibfk_2` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `parameter`
ADD CONSTRAINT `parameter_ibfk_1` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `recipient`
ADD CONSTRAINT `recipient_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `sender`
ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`),
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1,
ADD CONSTRAINT `sender_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

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
  (16, 'save-profiles'),
  (17, 'accept-campaign'),
  (18, 'get-log-main'),
  (19, 'get-log-api'),
  (20, 'get-log-campaign'),
  (21, 'get-log-statistic');

INSERT INTO `auth_user` (`id`, `auth_unit_id`, `name`, `password`) VALUES
  (1, 0, 'admin', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918'),
  (2, 1, 'user', '04f8996da763b7a969b1028ee3007569eaf3a635486ddab211d512c85b9df8fb');

INSERT INTO `auth_unit_right` (`id`, `auth_unit_id`, `auth_right_id`) VALUES
  (1, 1, 1),
  (2, 1, 4),
  (3, 1, 7),
  (4, 1, 9),
  (5, 1, 10),
  (6, 1, 17);