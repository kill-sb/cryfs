-- MySQL dump 10.19  Distrib 10.3.28-MariaDB, for Linux (x86_64)
--
-- Host: localhost    Database: cmit
-- ------------------------------------------------------
-- Server version	10.3.28-MariaDB

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `contacts`
--

DROP TABLE IF EXISTS `contacts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `contacts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(11) NOT NULL,
  `contactuserid` int(11) NOT NULL,
  `crtime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `userid` (`userid`) USING BTREE,
  KEY `contactuserid` (`contactuserid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `contacts`
--

LOCK TABLES `contacts` WRITE;
/*!40000 ALTER TABLE `contacts` DISABLE KEYS */;
INSERT INTO `contacts` VALUES (1,1,2,'2022-03-22 07:11:04'),(2,1,3,'2022-03-22 07:11:04'),(7,2,1,'2022-06-16 09:24:47'),(8,2,3,'2022-06-16 09:24:47');
/*!40000 ALTER TABLE `contacts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `efilemeta`
--

DROP TABLE IF EXISTS `efilemeta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `efilemeta` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) DEFAULT NULL,
  `descr` varchar(100) DEFAULT '',
  `fromrcid` int(11) NOT NULL DEFAULT -1,
  `ownerid` int(11) DEFAULT NULL,
  `crtime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `isdir` int(11) DEFAULT 0,
  `orgname` varchar(255) DEFAULT NULL,
  `multisrc` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=84 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `efilemeta`
--

LOCK TABLES `efilemeta` WRITE;
/*!40000 ALTER TABLE `efilemeta` DISABLE KEYS */;
INSERT INTO `efilemeta` VALUES (53,'cb2d7086-ff7d-4a2e-aed7-d545d27f5dce','cmit encrypted data',0,1,'2022-01-25 08:09:52',0,'Makefile',0),(54,'5a74f943-bbf8-475d-9135-434892b8970e','cmit encrypted data',0,1,'2022-01-25 09:20:10',0,'Makefile',0),(55,'78a24e23-86a9-4314-acfb-55400b196fcf','cmit encrypted data',0,1,'2022-01-26 07:19:15',0,'Makefile',0),(56,'9673841d-b1e7-4395-84c9-2c7ab97be635','cmit encrypted data',0,1,'2022-01-27 09:15:37',0,'TODO',0),(57,'6129b420-59e3-4e87-b449-d95426938b48','cmit encrypted data',0,1,'2022-01-27 09:20:54',0,'TODO',0),(58,'a7f014d9-73fe-44e5-82a3-d8f76ebf7529','cmit encrypted data',0,1,'2022-01-27 09:22:53',0,'TODO',0),(59,'a660fbb8-35ef-4470-ac76-3a3550776b02','cmit encrypted data',0,1,'2022-01-27 09:23:55',0,'TODO',0),(60,'dc082df5-c36d-4044-bc9f-011c13e41e4f','cmit encrypted data',0,1,'2022-01-28 04:20:21',0,'TODO',0),(61,'ef5e29c3-0b0b-4c49-a157-c81d836e7571','cmit encrypted data',0,1,'2022-01-28 04:23:06',0,'TODO',0),(62,'ef5e29c3-0b0b-4c49-a157-c81d836e7571','cmit encrypted data',0,1,'2022-01-28 04:28:15',0,'TODO',0),(63,'1ca01ce0-5243-488c-bcd2-a0c0409e9153','cmit encrypted dir',0,1,'2022-01-28 04:38:05',1,'nls',0),(64,'1a45df84-c3c8-47ff-a34b-128f4806ec7f','cmit encrypted data',0,1,'2022-01-28 04:40:39',0,'TODO',0),(65,'de97f6fc-1d5d-420a-8b2a-8261329ff041','cmit encrypted dir',0,1,'2022-01-28 04:51:35',1,'nls',0),(66,'b4f620ac-065b-4a4b-8a99-e2b68fe3953f','cmit encrypted data',0,1,'2022-02-11 08:08:00',0,'Makefile',0),(67,'ac1ada36-c79e-4951-bb6a-c37cca8952c0','cmit encrypted dir',0,1,'2022-02-11 08:08:59',1,'samplefs',0),(68,'f6e63544-ebe2-4c83-bdc9-9a10baf2438e','cmit encrypted data',0,1,'2022-03-04 02:38:55',0,'Makefile',0),(69,'02c4c033-5ab0-4aa0-8edc-bd4590b9056e','cmit encrypted dir',1,2,'2022-03-01 08:12:04',1,'1.csd.outdata',0),(71,'e9002c53-3a67-4a5b-8010-89a9a8e6dc06','test new data',1,1,'2022-04-06 07:25:25',0,'Makefile',0),(72,'1c6669f8-c74b-4522-886e-b19df1ec7f25','test new data',0,3,'2022-05-25 08:25:32',0,'newtests',0),(73,'e9002c53-3a67-4a5b-8010-89a9a8e6dc06','test new data',3,4,'2022-05-25 08:33:06',0,'Makefile',0),(74,'2c9f3bd3-db39-43da-b464-035dee54551d','test new data',3,4,'2022-05-25 08:39:28',0,'Makefile',0),(75,'7f5d4f25-3e43-4463-ab48-1ca5115c0ff4','cmit encrypted data',0,1,'2022-06-08 06:52:24',0,'Makefile',0),(76,'7109446b-73ab-4ae9-91ca-5eba7403b430','cmit encrypted dir',0,2,'2022-06-09 01:52:11',1,'sql',0),(77,'73c997d9-12cb-4610-82d6-cdbf2d47589f','cmit encrypted data',0,1,'2022-06-29 06:01:40',0,'contacts.sql',0),(78,'fb457b85-029b-49c8-ae36-5c5492bc6a7f','cmit encrypted dir',0,1,'2022-06-29 06:02:38',1,'backend',0),(79,'41778bf8-10f6-4c2c-ac81-98cbc168b542','cmit encrypted data',0,2,'2022-06-29 06:03:18',0,'cert.pem',0),(80,'a190cbb5-0fc5-47f1-8037-25b6cbac3fc2','cmit encrypted dir',0,2,'2022-06-29 06:25:42',1,'a190cbb5-0fc5-47f1-8037-25b6cbac3fc2',0),(81,'0e317b6d-20de-424e-a6a1-65f967424566','cmit encrypted dir',0,2,'2022-06-29 07:30:54',1,'0e317b6d-20de-424e-a6a1-65f967424566',0),(82,'1b452233-fb6b-4dcc-933d-f845b682f937','cmit encrypted data',0,2,'2022-06-29 07:39:16',0,'cert.pem',0),(83,'3e543971-fe8b-4ba6-afa6-ac2c5b484561','cmit encrypted dir',8,2,'2022-06-29 07:47:18',1,'3e543971-fe8b-4ba6-afa6-ac2c5b484561',0);
/*!40000 ALTER TABLE `efilemeta` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `expinvolvedata`
--

DROP TABLE IF EXISTS `expinvolvedata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `expinvolvedata` (
  `expid` int(11) NOT NULL,
  `datauuid` char(36) NOT NULL,
  `datatype` int(11) NOT NULL,
  `dataowner` int(11) NOT NULL,
  `nodeid` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `expinvolvedata`
--

LOCK TABLES `expinvolvedata` WRITE;
/*!40000 ALTER TABLE `expinvolvedata` DISABLE KEYS */;
INSERT INTO `expinvolvedata` VALUES (8,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,1),(9,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,2),(11,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,3),(13,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,7),(13,'f6e63544-ebe2-4c83-bdc9-9a10baf2438e',0,1,7),(13,'1c6669f8-c74b-4522-886e-b19df1ec7f25',0,3,8),(15,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,11),(15,'f6e63544-ebe2-4c83-bdc9-9a10baf2438e',0,1,11),(15,'1c6669f8-c74b-4522-886e-b19df1ec7f25',0,3,12),(16,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0,1,13),(17,'fb457b85-029b-49c8-ae36-5c5492bc6a7f',0,1,14),(18,'fb457b85-029b-49c8-ae36-5c5492bc6a7f',0,1,15),(19,'fb457b85-029b-49c8-ae36-5c5492bc6a7f',0,1,16),(20,'fb457b85-029b-49c8-ae36-5c5492bc6a7f',0,1,17);
/*!40000 ALTER TABLE `expinvolvedata` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `exports`
--

DROP TABLE IF EXISTS `exports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `exports` (
  `expid` int(11) NOT NULL AUTO_INCREMENT,
  `requid` int(11) NOT NULL,
  `status` int(11) NOT NULL,
  `datatype` int(11) NOT NULL,
  `datauuid` char(36) NOT NULL,
  `crtime` datetime NOT NULL,
  `comment` varchar(200) DEFAULT '',
  PRIMARY KEY (`expid`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `exports`
--

LOCK TABLES `exports` WRITE;
/*!40000 ALTER TABLE `exports` DISABLE KEYS */;
INSERT INTO `exports` VALUES (1,1,2,1,'1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3','2022-05-22 09:55:20',''),(2,4,2,1,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc','2022-05-22 09:58:41',''),(3,3,2,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-22 10:01:20',''),(4,3,2,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-22 10:07:03',''),(7,3,2,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-22 10:20:12',''),(8,3,-1,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-22 10:23:16',''),(9,3,2,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-22 10:53:57',''),(10,1,2,1,'1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3','2022-05-24 00:00:00',''),(11,3,1,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-24 00:00:00',''),(13,4,-1,1,'2c9f3bd3-db39-43da-b464-035dee545511','2022-05-25 16:54:44',''),(15,4,1,1,'2c9f3bd3-db39-43da-b464-035dee545511','2022-05-25 17:27:28',''),(16,3,2,1,'2e888dcf-7e68-4665-910a-eb906995b60f','2022-05-27 09:29:51',''),(17,3,2,1,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92','2022-06-29 15:03:56',''),(18,3,2,1,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92','2022-06-29 15:08:30','political order'),(19,3,2,1,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92','2022-06-29 15:22:03','international communication'),(20,3,2,1,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92','2022-06-29 15:26:05','test');
/*!40000 ALTER TABLE `exports` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `exprocque`
--

DROP TABLE IF EXISTS `exprocque`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `exprocque` (
  `expid` int(11) NOT NULL,
  `procuid` int(11) NOT NULL,
  `status` int(11) NOT NULL,
  `comment` varchar(200) DEFAULT '',
  `proctime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `nodeid` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`nodeid`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `exprocque`
--

LOCK TABLES `exprocque` WRITE;
/*!40000 ALTER TABLE `exprocque` DISABLE KEYS */;
INSERT INTO `exprocque` VALUES (8,1,-1,'piss off','2022-05-24 13:35:03',1),(9,2,2,'political order','2022-05-24 10:14:19',2),(11,1,1,'piss off','2022-05-27 02:58:39',3),(13,1,-1,'no way','2022-05-25 09:34:02',7),(13,3,2,'','2022-05-25 09:34:55',8),(15,1,1,'ok','2022-05-25 09:28:41',11),(15,3,1,'ok','2022-05-25 09:29:34',12),(16,1,2,'political order','2022-05-27 01:29:51',13),(17,1,2,'political order','2022-06-29 07:03:56',14),(18,1,2,'political order','2022-06-29 07:08:30',15),(19,1,2,'international communication','2022-06-29 07:22:03',16),(20,1,2,'test','2022-06-29 07:26:05',17);
/*!40000 ALTER TABLE `exprocque` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `notifies`
--

DROP TABLE IF EXISTS `notifies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `notifies` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type` int(11) NOT NULL,
  `content` varchar(4096) DEFAULT NULL,
  `descr` varchar(4096) DEFAULT '',
  `crtime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `fromuid` int(11) DEFAULT NULL,
  `touid` int(11) DEFAULT NULL,
  `isnew` int(11) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `notifies`
--

LOCK TABLES `notifies` WRITE;
/*!40000 ALTER TABLE `notifies` DISABLE KEYS */;
INSERT INTO `notifies` VALUES (6,2,'44acb473-1aa1-455f-b8a8-765d8d250d48','new shared data','2022-05-24 09:30:51',1,2,1),(7,2,'44acb473-1aa1-455f-b8a8-765d8d250d48','new shared data','2022-05-24 09:30:51',1,3,1),(8,2,'44acb473-1aa1-455f-b8a8-765d8d250d48','new shared data','2022-05-24 09:30:51',1,2,1),(9,3,'9','political order','2022-05-24 09:30:51',3,1,1),(10,3,'11','political order','2022-05-24 09:30:51',3,1,1),(11,1,'test a plain text','','2022-05-24 09:30:51',3,2,1),(12,1,'test a plain text','','2022-05-24 09:30:51',3,2,1),(13,3,'12','political order','2022-05-25 08:41:13',4,3,1),(14,3,'12','political order','2022-05-25 08:41:13',4,1,1),(15,3,'12','political order','2022-05-25 08:41:13',4,1,1),(16,3,'13','political order','2022-05-25 08:54:44',4,1,1),(17,3,'13','political order','2022-05-25 08:54:44',4,3,1),(18,3,'14','political order','2022-06-16 04:52:30',4,1,1),(19,3,'14','political order','2022-05-25 09:22:00',4,3,1),(20,3,'15','political order','2022-06-16 04:54:41',4,1,0),(21,3,'15','political order','2022-05-25 09:27:28',4,3,1),(23,3,'16','political order','2022-06-16 04:28:16',3,1,0),(24,0,'test a plain text','','2022-06-16 04:41:21',1,2,1),(25,3,'17','political order','2022-06-29 07:03:56',3,1,1),(26,3,'18','political order','2022-06-29 07:08:30',3,1,1),(27,3,'19','international communication','2022-06-29 07:22:03',3,1,1),(28,3,'20','test','2022-06-29 07:26:05',3,1,1);
/*!40000 ALTER TABLE `notifies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `rcimport`
--

DROP TABLE IF EXISTS `rcimport`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rcimport` (
  `rcid` int(11) DEFAULT NULL,
  `relname` varchar(4096) DEFAULT NULL,
  `filedesc` varchar(1024) DEFAULT NULL,
  `sha256` char(64) DEFAULT NULL,
  `size` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `rcimport`
--

LOCK TABLES `rcimport` WRITE;
/*!40000 ALTER TABLE `rcimport` DISABLE KEYS */;
INSERT INTO `rcimport` VALUES (1,'path/data1','data1','f6b0ff59bc0a97b8c293398546b001320ce3d127f32b23c1a5f562afdbf4c5c1',2048),(2,'path/data1','data1','f6b0ff59bc0a97b8c293398546b001320ce3d127f32b23c1a5f562afdbf4c5c1',2048),(3,'path/data1','data1','f6b0ff59bc0a97b8c293398546b001320ce3d127f32b23c1a5f562afdbf4c5c1',2048);
/*!40000 ALTER TABLE `rcimport` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `rcinputdata`
--

DROP TABLE IF EXISTS `rcinputdata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rcinputdata` (
  `rcid` int(11) DEFAULT NULL,
  `srcuuid` char(36) DEFAULT NULL,
  `srctype` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `rcinputdata`
--

LOCK TABLES `rcinputdata` WRITE;
/*!40000 ALTER TABLE `rcinputdata` DISABLE KEYS */;
INSERT INTO `rcinputdata` VALUES (1,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1),(1,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0),(2,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1),(2,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0),(3,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1),(3,'ac1ada36-c79e-4951-bb6a-c37cca8952c0',0),(3,'1c6669f8-c74b-4522-886e-b19df1ec7f25',0),(5,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',1),(5,'41778bf8-10f6-4c2c-ac81-98cbc168b542',0),(6,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',1),(6,'41778bf8-10f6-4c2c-ac81-98cbc168b542',0),(7,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',1),(7,'41778bf8-10f6-4c2c-ac81-98cbc168b542',0),(8,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',1),(8,'41778bf8-10f6-4c2c-ac81-98cbc168b542',0);
/*!40000 ALTER TABLE `rcinputdata` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `resetpasswordtokens`
--

DROP TABLE IF EXISTS `resetpasswordtokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `resetpasswordtokens` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(11) NOT NULL,
  `tokensha256` char(64) DEFAULT NULL,
  `crtime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `userid` (`userid`) USING BTREE,
  KEY `token` (`tokensha256`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `resetpasswordtokens`
--

LOCK TABLES `resetpasswordtokens` WRITE;
/*!40000 ALTER TABLE `resetpasswordtokens` DISABLE KEYS */;
/*!40000 ALTER TABLE `resetpasswordtokens` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `runcontext`
--

DROP TABLE IF EXISTS `runcontext`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `runcontext` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(11) NOT NULL,
  `os` varchar(32) DEFAULT '',
  `baseimg` varchar(128) DEFAULT '',
  `outputuuid` char(36) DEFAULT '',
  `crtime` datetime DEFAULT '2022-01-01 00:00:00',
  `detime` datetime DEFAULT '2022-01-01 00:00:00',
  `ipaddr` varchar(50) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `runcontext`
--

LOCK TABLES `runcontext` WRITE;
/*!40000 ALTER TABLE `runcontext` DISABLE KEYS */;
INSERT INTO `runcontext` VALUES (1,1,'linux','centos8','e9002c53-3a67-4a5b-8010-89a9a8e6dc06','2022-04-06 00:00:00','2022-04-06 15:00:01','192.168.80.138'),(2,1,'linux','centos8','','2022-04-06 00:00:00','2022-01-01 00:00:00','127.0.0.1'),(3,4,'linux','centos8','2c9f3bd3-db39-43da-b464-035dee54551d','2022-05-06 00:00:00','2022-05-26 15:00:01','127.0.0.1'),(4,2,'linux','cmit','7244de2f-0a62-4018-8c33-18e603eaac3d','2022-06-29 14:07:13','2022-06-29 14:07:44','127.0.0.1'),(5,2,'linux','cmit','a190cbb5-0fc5-47f1-8037-25b6cbac3fc2','2022-06-29 14:25:13','2022-06-29 14:25:42','127.0.0.1'),(6,2,'linux','cmit','0e317b6d-20de-424e-a6a1-65f967424566','2022-06-29 15:30:28','2022-06-29 15:30:54','127.0.0.1'),(7,2,'linux','cmit','1b452233-fb6b-4dcc-933d-f845b682f937','2022-06-29 15:38:58','2022-06-29 15:39:16','127.0.0.1'),(8,2,'linux','cmit','3e543971-fe8b-4ba6-afa6-ac2c5b484561','2022-06-29 15:47:06','2022-06-29 15:47:18','127.0.0.1');
/*!40000 ALTER TABLE `runcontext` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sharetags`
--

DROP TABLE IF EXISTS `sharetags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sharetags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) DEFAULT NULL,
  `ownerid` int(11) DEFAULT NULL,
  `receivers` text DEFAULT NULL,
  `expire` datetime DEFAULT NULL,
  `maxuse` int(11) DEFAULT NULL,
  `keycryptkey` char(32) DEFAULT NULL,
  `datauuid` char(36) DEFAULT NULL,
  `perm` int(11) DEFAULT NULL,
  `fromtype` int(11) DEFAULT NULL,
  `sha256` char(64) NOT NULL DEFAULT '',
  `crtime` datetime DEFAULT NULL,
  `orgname` varchar(255) DEFAULT '',
  `isdir` int(11) NOT NULL DEFAULT 0,
  `content` int(11) NOT NULL DEFAULT 0,
  `descr` varchar(255) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=66 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sharetags`
--

LOCK TABLES `sharetags` WRITE;
/*!40000 ALTER TABLE `sharetags` DISABLE KEYS */;
INSERT INTO `sharetags` VALUES (42,'5d12b326-3a7a-4727-a7bb-a2cace3450db',1,'li4,wang2','2999-12-31 00:00:00',-1,'08bb4f4e191a467770c86ba0ef57a4dc','1a45df84-c3c8-47ff-a34b-128f4806ec7f',1,0,'','2022-01-28 14:43:52','TODO',0,0,''),(43,'1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',2,'zhang3,li4,wang2','2999-12-31 00:00:00',-1,'ad5d314cf0496d46091f8a784f6270ab','5d12b326-3a7a-4727-a7bb-a2cace3450db',1,1,'','2022-01-28 15:40:43','TODO',0,0,''),(44,'773b4a93-aa9d-42a6-8b47-46bd89872832',1,'li4','2999-12-31 00:00:00',-1,'24dfbc58ec833a051c3ff1c22a6612b3','1a45df84-c3c8-47ff-a34b-128f4806ec7f',1,0,'','2022-01-30 09:20:32','TODO',0,0,''),(45,'c38eb51b-fea0-4654-9bfc-b1d4c83123cf',2,'wang2,cmit','2999-12-31 00:00:00',-1,'f473653760f74ce9108946711498d925','773b4a93-aa9d-42a6-8b47-46bd89872832',1,1,'','2022-01-30 09:21:42','TODO',0,0,''),(46,'44acb473-1aa1-455f-b8a8-765d8d250d48',1,'li4,wang2','2999-12-31 00:00:00',-1,'83a47e65d005760e499466cceba68eae','ac1ada36-c79e-4951-bb6a-c37cca8952c0',1,0,'','2022-02-11 16:09:32','',0,0,''),(47,'2e888dcf-7e68-4665-910a-eb906995b60f',2,'wang2','2999-12-31 00:00:00',-1,'162b4cb2cd7efbb4f564d34962c571af','44acb473-1aa1-455f-b8a8-765d8d250d48',1,1,'','2022-02-15 15:00:15','',0,0,''),(49,'b178ea6d-877d-405e-a440-35ba705773d1',3,'li4','2999-12-31 00:00:00',-1,'eb6586dcbf60bc559a1b31b4a6e496be','cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1,1,'','2022-03-02 14:00:00','Makefile',0,0,''),(50,'eb60fafe-bac1-4343-b863-5c60ddd7e5b8',2,'li4','2999-12-31 00:00:00',-1,'93ab627ed3433957130d639c564e0ad6','b178ea6d-877d-405e-a440-35ba705773d1',1,1,'','2022-03-02 14:35:11','Makefile',0,0,''),(52,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1,'wang2,cmit','2999-12-31 00:00:00',-1,'3b7b57bef298969fcf272b8caa56d3d6','f6e63544-ebe2-4c83-bdc9-9a10baf2438e',1,0,'','2022-01-30 09:21:42','TODO',0,0,''),(53,'2c9f3bd3-db39-43da-b464-035dee545511',4,'li4,wang2,zhang3','2999-12-31 00:00:00',-1,'3b7b57bef298969fcf272b8caa56d3d6','2c9f3bd3-db39-43da-b464-035dee54551d',1,0,'f6b0ff59bc0a97b8c293398546b001320ce3d127f32b23c1a5f562afdbf4c5c1','2022-05-25 09:21:42','Makefile',0,0,''),(54,'24acb54e-bb21-4aaf-8508-150debc91334',1,'li4,wang2','2999-12-31 00:00:00',-1,'376af53c0aaa011bbb945731be095252','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',1,0,'','2022-06-09 08:27:28','Makefile',0,0,''),(55,'128fff31-5a5f-47f1-b08b-35d6968f3768',1,'li4','2999-12-31 00:00:00',3,'f30f04dac481014ae4302fd7c4022d85','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',0,0,'','2022-06-09 09:08:09','Makefile',0,0,''),(56,'35d8ad97-23cb-46cc-88c0-cde84b2f82d1',1,'li4','2999-12-31 00:00:00',3,'33032d1d67d256cd9c9229aec6ad5e8e','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',1,0,'','2022-06-09 09:16:52','Makefile',0,0,''),(57,'408f188e-f35d-4c7d-9ccd-57d11d6d285e',1,'li4','2999-12-31 00:00:00',-1,'98f5ff8fef608e82dbddee905c1aa564','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',1,0,'','2022-06-09 09:27:32','Makefile',0,0,''),(58,'ab2b67ec-1a7e-4043-adee-cf5b8ccc88d6',1,'li4','2999-12-31 00:00:00',-1,'3f3abf2747bd67b8acb67461362f0d4e','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',1,0,'','2022-06-09 09:28:46','Makefile',0,0,''),(59,'906a5b2f-40c0-415b-9057-0083cb8a0b23',1,'li4','2999-12-31 00:00:00',-1,'dfb194f8752ea829178e72184ac77c40','7f5d4f25-3e43-4463-ab48-1ca5115c0ff4',1,0,'','2022-06-09 09:50:20','Makefile',0,0,''),(60,'fd588fdc-e83d-4832-9d1d-2d609c952bcb',2,'wang2','2999-12-31 00:00:00',-1,'ec8a66943ef0f96d90ab8211cd6941e0','7109446b-73ab-4ae9-91ca-5eba7403b430',1,0,'','2022-06-09 09:53:35','',1,0,''),(61,'3a4226e5-86f4-417f-96b4-ce88d117301a',2,'wang2','2999-12-31 00:00:00',3,'0daaa6717cce11902d908cb6322aed0e','7109446b-73ab-4ae9-91ca-5eba7403b430',1,0,'','2022-06-09 09:54:05','',1,0,''),(62,'f601753d-ec42-4d05-96cc-9799eba7c649',2,'wang2','2999-12-31 00:00:00',3,'8f7abfe3d7c7a74a28b27cc881bc56c0','7109446b-73ab-4ae9-91ca-5eba7403b430',1,0,'','2022-06-09 10:58:24','sql',1,0,''),(63,'9083f1ec-4972-4453-ae9f-c284d0abe249',3,'li4,zhang3','2999-12-31 00:00:00',-1,'b2ddb11d041de571c4d8516ec3c20c4c','f601753d-ec42-4d05-96cc-9799eba7c649',1,1,'','2022-06-09 11:02:14','sql',1,0,''),(64,'a1609420-20bf-48cb-895b-c65dc4170f51',2,'wang2','2999-12-31 00:00:00',2,'fc3d1bfa1eae15d7865d24c58f587580','906a5b2f-40c0-415b-9057-0083cb8a0b23',1,1,'','2022-06-09 11:04:56','Makefile',0,0,''),(65,'7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',1,'li4,wang2','2999-12-31 00:00:00',-1,'ba3ac6de65a9675999956b7dc643020b','fb457b85-029b-49c8-ae36-5c5492bc6a7f',1,0,'','2022-06-29 14:04:41','backend',1,0,'');
/*!40000 ALTER TABLE `sharetags` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `shareusers`
--

DROP TABLE IF EXISTS `shareusers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `shareusers` (
  `taguuid` char(36) NOT NULL,
  `userid` int(11) NOT NULL,
  `leftuse` int(11) DEFAULT -1
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `shareusers`
--

LOCK TABLES `shareusers` WRITE;
/*!40000 ALTER TABLE `shareusers` DISABLE KEYS */;
INSERT INTO `shareusers` VALUES ('5d12b326-3a7a-4727-a7bb-a2cace3450db',2,-1),('5d12b326-3a7a-4727-a7bb-a2cace3450db',3,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',1,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',2,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',3,-1),('773b4a93-aa9d-42a6-8b47-46bd89872832',2,-1),('c38eb51b-fea0-4654-9bfc-b1d4c83123cf',3,-1),('c38eb51b-fea0-4654-9bfc-b1d4c83123cf',4,-1),('44acb473-1aa1-455f-b8a8-765d8d250d48',2,-1),('44acb473-1aa1-455f-b8a8-765d8d250d48',0,-1),('2e888dcf-7e68-4665-910a-eb906995b60f',0,-1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',2,3),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',3,1),('e40fc6f2-ae6d-4278-9b64-612646aaa7d0',2,-1),('971f7c51-98f1-401d-a38f-517eb23621d5',2,-1),('b178ea6d-877d-405e-a440-35ba705773d1',2,-1),('eb60fafe-bac1-4343-b863-5c60ddd7e5b8',2,-1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',3,1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',4,-1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',3,1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',4,-1),('2c9f3bd3-db39-43da-b464-035dee545511',1,-1),('2c9f3bd3-db39-43da-b464-035dee545511',2,-1),('2c9f3bd3-db39-43da-b464-035dee545511',3,-1),('24acb54e-bb21-4aaf-8508-150debc91334',2,-1),('24acb54e-bb21-4aaf-8508-150debc91334',3,-1),('128fff31-5a5f-47f1-b08b-35d6968f3768',2,3),('35d8ad97-23cb-46cc-88c0-cde84b2f82d1',2,1),('408f188e-f35d-4c7d-9ccd-57d11d6d285e',2,-1),('ab2b67ec-1a7e-4043-adee-cf5b8ccc88d6',2,-1),('906a5b2f-40c0-415b-9057-0083cb8a0b23',2,-1),('fd588fdc-e83d-4832-9d1d-2d609c952bcb',3,-1),('3a4226e5-86f4-417f-96b4-ce88d117301a',3,0),('f601753d-ec42-4d05-96cc-9799eba7c649',3,1),('9083f1ec-4972-4453-ae9f-c284d0abe249',1,-1),('9083f1ec-4972-4453-ae9f-c284d0abe249',2,-1),('a1609420-20bf-48cb-895b-c65dc4170f51',3,0),('7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',2,-1),('7ff9419a-d54e-4f17-8baf-cd3d9a7d9c92',3,-1);
/*!40000 ALTER TABLE `shareusers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pwdsha256` char(64) NOT NULL,
  `descr` varchar(100) DEFAULT '',
  `enclocalkey` char(32) NOT NULL,
  `register` datetime DEFAULT '2022-02-02 00:00:00',
  `name` varchar(16) NOT NULL,
  `pubkey` varchar(1024) DEFAULT '',
  `mobile` varchar(20) DEFAULT '',
  `email` varchar(50) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92','zhang3','f447b20a7fcbf53a5d5be013ea0b15af','2021-11-05 11:05:00','zhang3','','139','a@a.com'),(2,'96cae35ce8a9b0244178bf28e4966c2ce1b8385723a96a6b838858cdd6ca0a1e','li 4','4297f44b13955235245b2497399d7a93','2021-11-06 16:01:00','li4','','13811111111','li4@a.com'),(3,'481f6cc0511143ccdd7e2d1b1b94faf0a700a8b49cd13922a70b5ae28acaa8c5','','4a62cf6ee3f8d889e65af1cc271f20fa','2021-11-19 09:35:00','wang2','','',''),(4,'dd007f90f6a6f9f2b15ca4afec79e3465fa4ad0a14bd590a1d2c18abeedcb410','','c36f10fd0ff59c3bcce088d7a7a6c410','2022-02-02 00:00:00','cmit','','13999999999','cmitfs@cmit.com');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2022-06-29 16:09:56
