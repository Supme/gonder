SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

CREATE TABLE IF NOT EXISTS `from` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

ALTER TABLE `from`
 ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`);

ALTER TABLE `from`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

ALTER TABLE `from`
ADD CONSTRAINT `from_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `campaign` ADD `from_id` int(11) NOT NULL DEFAULT '0' ;

ALTER TABLE `campaign` ADD KEY `from_id` (`from_id`);