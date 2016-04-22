SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

ALTER TABLE `campaign` CHANGE `send_unsubscribe` `send_unsubscribe` TINYINT(1) NOT NULL DEFAULT '0';
ALTER TABLE `campaign` ADD `accepted` TINYINT(1) NOT NULL DEFAULT '0' ;

CREATE TABLE IF NOT EXISTS `auth_right` (
  `id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_unit` (
  `id` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_unit_right` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `auth_right_id` int(11) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_user` (
  `id` int(11) NOT NULL,
  `auth_unit_id` int(11) NOT NULL,
  `name` text NOT NULL,
  `password` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `auth_user_group` (
  `id` int(11) NOT NULL,
  `auth_user_id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

ALTER TABLE `auth_right`
ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `id` (`id`), ADD KEY `id_2` (`id`);

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

ALTER TABLE `auth_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `auth_unit`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `auth_unit_right`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `auth_unit_right`
ADD CONSTRAINT `auth_unit_right_ibfk_1` FOREIGN KEY (`auth_unit_id`) REFERENCES `auth_unit` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `auth_unit_right_ibfk_2` FOREIGN KEY (`auth_right_id`) REFERENCES `auth_right` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `auth_user`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `auth_user_group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `auth_user_group`
ADD CONSTRAINT `auth_user_group_ibfk_1` FOREIGN KEY (`auth_user_id`) REFERENCES `auth_user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `auth_user_group_ibfk_2` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

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
  (17, 'accept-campaign');

INSERT INTO `auth_unit` (`id`, `name`) VALUES
  (0, 'administrator'),
  (1, 'accepter');

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