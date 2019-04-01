CREATE TABLE `question` (`id` int(11) NOT NULL, `recipient_id` int(11) NOT NULL, `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
ALTER TABLE `question` ADD PRIMARY KEY (`id`), ADD KEY `recipient_id` (`recipient_id`);
ALTER TABLE `question` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `question` ADD CONSTRAINT `question_ibfk_1` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

CREATE TABLE `question_data` (`id` int(11) NOT NULL, `question_id` int(11) NOT NULL, `name` varchar(100) NOT NULL, `value` text NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
ALTER TABLE `question_data` ADD PRIMARY KEY (`id`), ADD KEY `question_id` (`question_id`);
ALTER TABLE `question_data` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `question_data` ADD CONSTRAINT `question_data_ibfk_1` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
