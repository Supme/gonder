ALTER TABLE `attachment` DROP `file`;

ALTER TABLE `sender`
  ADD `dkim_selector` VARCHAR(20) NOT NULL AFTER `name`,
  ADD `dkim_key` VARCHAR(2000) NOT NULL AFTER `dkim_selector`,
  ADD `dkim_use` BOOLEAN NOT NULL AFTER `dkim_key`;