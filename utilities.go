package main

import (
	"database/sql"
	"log"
)

const (
	tblChannels = "CREATE TABLE IF NOT EXISTS `channels` (  `id` varchar(18) NOT NULL,  `name` text NOT NULL DEFAULT '',  `serverId` varchar(18) NOT NULL,  PRIMARY KEY (`id`),  KEY `FK_channels_server` (`serverId`),  CONSTRAINT `FK_channels_server` FOREIGN KEY (`serverId`) REFERENCES `server` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	tblPings    = "CREATE TABLE IF NOT EXISTS `pings` (  `id` int(11) NOT NULL AUTO_INCREMENT,  `menzionatoreId` varchar(18) NOT NULL,  `menzionatoId` varchar(18) NOT NULL,  `channelId` varchar(18) NOT NULL,  `serverId` varchar(18) NOT NULL,  `timestamp` datetime NOT NULL,  `deleted` tinyint(1) NOT NULL DEFAULT 0,  `messageId` varchar(18) NOT NULL,  PRIMARY KEY (`id`),  UNIQUE KEY `messageId` (`messageId`),  KEY `FK_pings_channels` (`channelId`),  KEY `FK_pings_server` (`serverId`),  KEY `FK_pings_users` (`menzionatoreId`),  KEY `FK_pings_users_2` (`menzionatoId`),  CONSTRAINT `FK_pings_channels` FOREIGN KEY (`channelId`) REFERENCES `channels` (`id`),  CONSTRAINT `FK_pings_server` FOREIGN KEY (`serverId`) REFERENCES `server` (`id`),  CONSTRAINT `FK_pings_users` FOREIGN KEY (`menzionatoreId`) REFERENCES `users` (`id`),  CONSTRAINT `FK_pings_users_2` FOREIGN KEY (`menzionatoId`) REFERENCES `users` (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	tblServers  = "CREATE TABLE IF NOT EXISTS `server` (  `id` varchar(18) NOT NULL,  `name` varchar(100) NOT NULL,  PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	tblUsers    = "CREATE TABLE IF NOT EXISTS `users` (  `id` varchar(18) NOT NULL,  `nickname` varchar(32) NOT NULL,  PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
)

func execQuery(query string, db *sql.DB) {
	statement, err := db.Prepare(query)
	if err != nil {
		log.Println("Error preparing query,", err)
		return
	}

	_, err = statement.Exec()
	if err != nil {
		log.Println("Error creating table,", err)
	}
}
