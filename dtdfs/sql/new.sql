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
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `contacts`
--

LOCK TABLES `contacts` WRITE;
/*!40000 ALTER TABLE `contacts` DISABLE KEYS */;
INSERT INTO `contacts` VALUES (1,1,2,'2022-03-22 07:11:04'),(2,1,3,'2022-03-22 07:11:04');
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
) ENGINE=InnoDB AUTO_INCREMENT=71 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `efilemeta`
--

LOCK TABLES `efilemeta` WRITE;
/*!40000 ALTER TABLE `efilemeta` DISABLE KEYS */;
INSERT INTO `efilemeta` VALUES (53,'cb2d7086-ff7d-4a2e-aed7-d545d27f5dce','cmit encrypted data',0,1,'2022-01-25 08:09:52',0,'Makefile',0),(54,'5a74f943-bbf8-475d-9135-434892b8970e','cmit encrypted data',0,1,'2022-01-25 09:20:10',0,'Makefile',0),(55,'78a24e23-86a9-4314-acfb-55400b196fcf','cmit encrypted data',0,1,'2022-01-26 07:19:15',0,'Makefile',0),(56,'9673841d-b1e7-4395-84c9-2c7ab97be635','cmit encrypted data',0,1,'2022-01-27 09:15:37',0,'TODO',0),(57,'6129b420-59e3-4e87-b449-d95426938b48','cmit encrypted data',0,1,'2022-01-27 09:20:54',0,'TODO',0),(58,'a7f014d9-73fe-44e5-82a3-d8f76ebf7529','cmit encrypted data',0,1,'2022-01-27 09:22:53',0,'TODO',0),(59,'a660fbb8-35ef-4470-ac76-3a3550776b02','cmit encrypted data',0,1,'2022-01-27 09:23:55',0,'TODO',0),(60,'dc082df5-c36d-4044-bc9f-011c13e41e4f','cmit encrypted data',0,1,'2022-01-28 04:20:21',0,'TODO',0),(61,'ef5e29c3-0b0b-4c49-a157-c81d836e7571','cmit encrypted data',0,1,'2022-01-28 04:23:06',0,'TODO',0),(62,'ef5e29c3-0b0b-4c49-a157-c81d836e7571','cmit encrypted data',0,1,'2022-01-28 04:28:15',0,'TODO',0),(63,'1ca01ce0-5243-488c-bcd2-a0c0409e9153','cmit encrypted dir',0,1,'2022-01-28 04:38:05',1,'nls',0),(64,'1a45df84-c3c8-47ff-a34b-128f4806ec7f','cmit encrypted data',0,1,'2022-01-28 04:40:39',0,'TODO',0),(65,'de97f6fc-1d5d-420a-8b2a-8261329ff041','cmit encrypted dir',0,1,'2022-01-28 04:51:35',1,'nls',0),(66,'b4f620ac-065b-4a4b-8a99-e2b68fe3953f','cmit encrypted data',0,1,'2022-02-11 08:08:00',0,'Makefile',0),(67,'ac1ada36-c79e-4951-bb6a-c37cca8952c0','cmit encrypted dir',0,1,'2022-02-11 08:08:59',1,'samplefs',0),(68,'f6e63544-ebe2-4c83-bdc9-9a10baf2438e','cmit encrypted data',0,1,'2022-03-04 02:38:55',0,'Makefile',0),(69,'02c4c033-5ab0-4aa0-8edc-bd4590b9056e','cmit encrypted dir',1,2,'2022-03-01 08:12:04',1,'1.csd.outdata',0),(70,'e9002c53-3a67-4a5b-8010-89a9a8e6dc06','cmit encrypted data',1,2,'2022-03-09 08:16:53',0,'Makefile',0);
/*!40000 ALTER TABLE `efilemeta` ENABLE KEYS */;
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
  `os` varchar(32) DEFAULT NULL,
  `baseimg` varchar(128) DEFAULT NULL,
  `outputuuid` char(36) DEFAULT '',
  `crtime` datetime DEFAULT '2022-01-01 00:00:00',
  `detime` datetime DEFAULT '2022-01-01 00:00:00',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `runcontext`
--

LOCK TABLES `runcontext` WRITE;
/*!40000 ALTER TABLE `runcontext` DISABLE KEYS */;
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
  `hashmd5` char(32) DEFAULT '',
  `crtime` datetime DEFAULT NULL,
  `orgname` varchar(255) DEFAULT '',
  `isdir` int(11) NOT NULL DEFAULT 0,
  `content` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sharetags`
--

LOCK TABLES `sharetags` WRITE;
/*!40000 ALTER TABLE `sharetags` DISABLE KEYS */;
INSERT INTO `sharetags` VALUES (42,'5d12b326-3a7a-4727-a7bb-a2cace3450db',1,'li4,wang2','2999-12-31 00:00:00',-1,'08bb4f4e191a467770c86ba0ef57a4dc','1a45df84-c3c8-47ff-a34b-128f4806ec7f',1,0,NULL,'2022-01-28 14:43:52','TODO',0,0),(43,'1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',2,'zhang3,li4,wang2','2999-12-31 00:00:00',-1,'ad5d314cf0496d46091f8a784f6270ab','5d12b326-3a7a-4727-a7bb-a2cace3450db',1,1,NULL,'2022-01-28 15:40:43','TODO',0,0),(44,'773b4a93-aa9d-42a6-8b47-46bd89872832',1,'li4','2999-12-31 00:00:00',-1,'24dfbc58ec833a051c3ff1c22a6612b3','1a45df84-c3c8-47ff-a34b-128f4806ec7f',1,0,NULL,'2022-01-30 09:20:32','TODO',0,0),(45,'c38eb51b-fea0-4654-9bfc-b1d4c83123cf',2,'wang2,cmit','2999-12-31 00:00:00',-1,'f473653760f74ce9108946711498d925','773b4a93-aa9d-42a6-8b47-46bd89872832',1,1,NULL,'2022-01-30 09:21:42','TODO',0,0),(46,'44acb473-1aa1-455f-b8a8-765d8d250d48',1,'li4,wang2','2999-12-31 00:00:00',-1,'83a47e65d005760e499466cceba68eae','ac1ada36-c79e-4951-bb6a-c37cca8952c0',1,0,NULL,'2022-02-11 16:09:32','',0,0),(47,'2e888dcf-7e68-4665-910a-eb906995b60f',2,'wang2','2999-12-31 00:00:00',-1,'162b4cb2cd7efbb4f564d34962c571af','44acb473-1aa1-455f-b8a8-765d8d250d48',1,1,NULL,'2022-02-15 15:00:15','',0,0),(48,'cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1,'li4,wang2','2999-12-31 00:00:00',3,'3b7b57bef298969fcf272b8caa56d3d6','f6e63544-ebe2-4c83-bdc9-9a10baf2438e',1,0,'','2022-03-01 16:07:57','Makefile',0,0),(49,'b178ea6d-877d-405e-a440-35ba705773d1',3,'li4','2999-12-31 00:00:00',-1,'eb6586dcbf60bc559a1b31b4a6e496be','cd77acc1-3a40-4e02-8ddc-acc7a67474cc',1,1,'','2022-03-02 14:00:00','Makefile',0,0),(50,'eb60fafe-bac1-4343-b863-5c60ddd7e5b8',2,'li4','2999-12-31 00:00:00',-1,'93ab627ed3433957130d639c564e0ad6','b178ea6d-877d-405e-a440-35ba705773d1',1,1,'','2022-03-02 14:35:11','Makefile',0,0);
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
INSERT INTO `shareusers` VALUES ('5d12b326-3a7a-4727-a7bb-a2cace3450db',2,-1),('5d12b326-3a7a-4727-a7bb-a2cace3450db',3,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',1,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',2,-1),('1a6c3ece-dc57-4cc9-ae3c-d7f50d9362e3',3,-1),('773b4a93-aa9d-42a6-8b47-46bd89872832',2,-1),('c38eb51b-fea0-4654-9bfc-b1d4c83123cf',3,-1),('c38eb51b-fea0-4654-9bfc-b1d4c83123cf',4,-1),('44acb473-1aa1-455f-b8a8-765d8d250d48',2,-1),('44acb473-1aa1-455f-b8a8-765d8d250d48',0,-1),('2e888dcf-7e68-4665-910a-eb906995b60f',0,-1),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',2,3),('cd77acc1-3a40-4e02-8ddc-acc7a67474cc',3,2),('e40fc6f2-ae6d-4278-9b64-612646aaa7d0',2,-1),('971f7c51-98f1-401d-a38f-517eb23621d5',2,-1),('b178ea6d-877d-405e-a440-35ba705773d1',2,-1),('eb60fafe-bac1-4343-b863-5c60ddd7e5b8',2,-1);
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

-- Dump completed on 2022-03-28 15:26:54
