CREATE TABLE `unsubscribe_extra` (
  `id` int(11) NOT NULL,
  `unsubscribe_id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `unsubscribe_extra`
  ADD PRIMARY KEY (`id`),
  ADD KEY `unsubscribe_id` (`unsubscribe_id`) USING BTREE;

ALTER TABLE `unsubscribe_extra`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

ALTER TABLE `unsubscribe_extra`
  ADD CONSTRAINT `unsubscribe_extra_ibfk_1` FOREIGN KEY (`unsubscribe_id`) REFERENCES `unsubscribe` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
