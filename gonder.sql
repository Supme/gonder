SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- База данных: `gosender`
--

-- --------------------------------------------------------

--
-- Структура таблицы `attachment`
--

CREATE TABLE IF NOT EXISTS `attachment` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `path` text NOT NULL,
  `file` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `campaign`
--

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
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `group`
--

CREATE TABLE IF NOT EXISTS `group` (
`id` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `interface`
--

CREATE TABLE IF NOT EXISTS `interface` (
`id` int(11) NOT NULL,
  `name` text NOT NULL,
  `iface` text NOT NULL COMMENT 'Example: xx.xx.xx.xx or socks://ip:port for socks5. Blank for default interface.',
  `host` text NOT NULL COMMENT 'The name of the server on behalf of which there is a sending.',
  `stream` int(11) NOT NULL COMMENT 'Number of concurrent streams',
  `delay` int(11) NOT NULL COMMENT 'Delay between one time stream in seconds'
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `jumping`
--

CREATE TABLE IF NOT EXISTS `jumping` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `url` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `parameter`
--

CREATE TABLE IF NOT EXISTS `parameter` (
`id` int(11) NOT NULL,
  `recipient_id` int(11) NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=435 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `recipient`
--

CREATE TABLE IF NOT EXISTS `recipient` (
`id` int(11) NOT NULL,
  `campaign_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `name` text NOT NULL,
  `status` text,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB AUTO_INCREMENT=145771 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `unsubscribe`
--

CREATE TABLE IF NOT EXISTS `unsubscribe` (
`id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `email` text NOT NULL,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `attachment`
--
ALTER TABLE `attachment`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

--
-- Индексы таблицы `campaign`
--
ALTER TABLE `campaign`
 ADD PRIMARY KEY (`id`), ADD KEY `interface_id` (`interface_id`), ADD KEY `group_id` (`group_id`);

--
-- Индексы таблицы `group`
--
ALTER TABLE `group`
 ADD PRIMARY KEY (`id`);

--
-- Индексы таблицы `interface`
--
ALTER TABLE `interface`
 ADD PRIMARY KEY (`id`);

--
-- Индексы таблицы `jumping`
--
ALTER TABLE `jumping`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`), ADD KEY `recipient_id` (`recipient_id`);

--
-- Индексы таблицы `parameter`
--
ALTER TABLE `parameter`
 ADD PRIMARY KEY (`id`), ADD KEY `recipient_id` (`recipient_id`);

--
-- Индексы таблицы `recipient`
--
ALTER TABLE `recipient`
 ADD PRIMARY KEY (`id`), ADD KEY `campaign_id` (`campaign_id`);

--
-- Индексы таблицы `unsubscribe`
--
ALTER TABLE `unsubscribe`
 ADD PRIMARY KEY (`id`), ADD KEY `group_id` (`group_id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `attachment`
--
ALTER TABLE `attachment`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=2;
--
-- AUTO_INCREMENT для таблицы `campaign`
--
ALTER TABLE `campaign`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=5;
--
-- AUTO_INCREMENT для таблицы `group`
--
ALTER TABLE `group`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=3;
--
-- AUTO_INCREMENT для таблицы `interface`
--
ALTER TABLE `interface`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=5;
--
-- AUTO_INCREMENT для таблицы `jumping`
--
ALTER TABLE `jumping`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=5;
--
-- AUTO_INCREMENT для таблицы `parameter`
--
ALTER TABLE `parameter`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=435;
--
-- AUTO_INCREMENT для таблицы `recipient`
--
ALTER TABLE `recipient`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=145771;
--
-- AUTO_INCREMENT для таблицы `unsubscribe`
--
ALTER TABLE `unsubscribe`
MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `attachment`
--
ALTER TABLE `attachment`
ADD CONSTRAINT `attachment_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ограничения внешнего ключа таблицы `jumping`
--
ALTER TABLE `jumping`
ADD CONSTRAINT `jumping_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
ADD CONSTRAINT `jumping_ibfk_2` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ограничения внешнего ключа таблицы `parameter`
--
ALTER TABLE `parameter`
ADD CONSTRAINT `parameter_ibfk_1` FOREIGN KEY (`recipient_id`) REFERENCES `recipient` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ограничения внешнего ключа таблицы `recipient`
--
ALTER TABLE `recipient`
ADD CONSTRAINT `recipient_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaign` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
