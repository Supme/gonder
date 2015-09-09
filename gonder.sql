SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;


CREATE TABLE IF NOT EXISTS `attachment` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `path` text NOT NULL,
  `file` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `campaign` (
`id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `interface_id` int(11) NOT NULL,
  `from` text NOT NULL,
  `from_name` text NOT NULL,
  `name` text NOT NULL,
  `subject` text NOT NULL,
  `message` text NOT NULL,
  `start_time` timestamp NULL DEFAULT NULL,
  `end_time` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `group` (
`id` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `interface` (
`id` int(11) NOT NULL,
  `name` text NOT NULL,
  `iface` text NOT NULL COMMENT 'Example: eth0 or xx.xx.xx.xx:xxxx for socks5. Blank for default interface.',
  `host` text NOT NULL COMMENT 'The name of the server on behalf of which there is a sending.',
  `stream` int(11) NOT NULL COMMENT 'Number of concurrent streams',
  `delay` int(11) NOT NULL COMMENT 'Delay between one time stream in seconds'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `jumping` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `url` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `parameter` (
`id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `recipient` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  `status` text,
  `opened` int(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `unsubscribe` (
`id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


ALTER TABLE `attachment`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `campaign`
 ADD PRIMARY KEY (`id`), ADD KEY `interface_id` (`interface_id`), ADD KEY `group_id` (`group_id`);

ALTER TABLE `group`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `interface`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `jumping`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`), ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `parameter`
 ADD PRIMARY KEY (`id`), ADD KEY `recipient_id` (`recipient_id`);

ALTER TABLE `recipient`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

ALTER TABLE `unsubscribe`
 ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`);


ALTER TABLE `attachment`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `campaign`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `interface`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `jumping`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `parameter`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `recipient`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `unsubscribe`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

ALTER TABLE `attachment`
ADD CONSTRAINT `attachment_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `jumping`
ADD CONSTRAINT `jumping_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `jumping_ibfk_2` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `parameter`
ADD CONSTRAINT `parameter_ibfk_1` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `recipient`
ADD CONSTRAINT `recipient_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;