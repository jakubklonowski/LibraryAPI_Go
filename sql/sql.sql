-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Wersja serwera:               10.9.2-MariaDB - mariadb.org binary distribution
-- Serwer OS:                    Win64
-- HeidiSQL Wersja:              11.3.0.6295
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- Zrzut struktury bazy danych library
CREATE DATABASE IF NOT EXISTS `library` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;
USE `library`;

-- Zrzut struktury tabela library.book
CREATE TABLE IF NOT EXISTS `book` (
  `ID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `Name` varchar(50) NOT NULL,
  `Author` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Eksport danych został odznaczony.

-- Zrzut struktury tabela library.client
CREATE TABLE IF NOT EXISTS `client` (
  `ID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `Name` varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Eksport danych został odznaczony.

-- Zrzut struktury tabela library.library
CREATE TABLE IF NOT EXISTS `library` (
  `ID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ID_Book` int(10) unsigned NOT NULL,
  `ID_Client` int(10) unsigned NOT NULL,
  `Date` datetime NOT NULL DEFAULT current_timestamp(),
  `Active` tinyint(4) NOT NULL DEFAULT 1,
  PRIMARY KEY (`ID`),
  KEY `Kolumna 2` (`ID_Book`),
  KEY `Kolumna 3` (`ID_Client`),
  CONSTRAINT `FK_Library_Book` FOREIGN KEY (`ID_Book`) REFERENCES `book` (`ID`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_Library_Client` FOREIGN KEY (`ID_Client`) REFERENCES `client` (`ID`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Table for borrowed books.';

-- Eksport danych został odznaczony.

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
