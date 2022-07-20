#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>
#include <sys/stat.h>


#define TMPFILE "/tmp/.dtdfs.tgz"
#define TMPDIR "/tmp/.dtdfs_files"
#define INSTALL_DIR "/usr/bin"
#define UBT_CERT "/usr/local/share/ca-certificates/apisvr_cert.pem"

#define SKIP_BYTE 65536

void GetSelf(char *path)
{
	char link[4096];
	sprintf(link,"/proc/%d/exe",getpid());
	readlink(link,path,4095);
}

int Installed(){
	if(system("dtdfs -h >/dev/null 2>/dev/null")==0){
		return 1;
	}
	return 0;
}

int UseUbtCfg() // return 1, treat as ubuntu >22.04, 0 not ubuntu, -1 ubuntu  but low edition
{
	char os[1024],ver[1024];
	FILE *fp=fopen("/etc/issue","r");
	if (!fp)
		return 0;
	fscanf(fp,"%s%s",os,ver);
	fclose(fp);
	if (strncasecmp(os,"ubuntu",6)==0){
		if(strcmp(ver,"22.04")>=0)
			return 1;
		else 	
			return -1;
	}else 
		return 0;
}

void Uninstall(int useubt)
{
	char cmd[1024];
	printf("OK\nUninstalling...");
	system("sed -i '/ apisvr /d' /etc/hosts >/dev/null 2>/dev/null");
	system("podman rmi cmit >/dev/null 2>/dev/null");
	sprintf(cmd,"rm -f %s/dtdfs %s/datamgr %s/cmfs >/dev/null 2>/dev/null",INSTALL_DIR,INSTALL_DIR,INSTALL_DIR);
	system(cmd);
	if(useubt){
			system("rm -f "UBT_CERT" >/dev/null 2>/dev/null");
			system("update-ca-certificates >/dev/null 2>/dev/null");
	}
	printf("OK\nData Defense linux client has been uninstalled from your system\n");
}

int main(int c, char** v)
{
	char bin[2048];
	int docker;
	char IP[512];
	char cmd[4096];
	int ins=0;
	int useubt=0;

	printf("Checking environment...");
	fflush(stdout);
	useubt=UseUbtCfg();
	if (useubt<0){
			printf("FAILED\nUbuntu system should newer than 22.04\n");
			exit(1);
	}
	ins=Installed();
	// unistall
	if (c==2 && strcmp(v[1],"-u")==0){
		if (ins)
			Uninstall(useubt);
		else 
			printf("FAILED\nData Defense is not found in your system\n");
		exit(0);
	}

	// install start
	if (ins){
		printf("FAILED\nData Defense has already been installed\n");
		exit(1);
	}

	docker=system("podman -v 1>/dev/null 2>/dev/null");
	if (docker!=0){
		printf("FAILED\nPodman not found, if you are using Centos 8.x, type 'yum install podman-docker' to install it, if you are using other OS, try to use dnf/yum/apt  install podman packges first.\n");
		exit(1);
	}

	if (c>2 && strcmp(v[1],"-svr")==0){
		if (strlen(v[2])<256)
			strcpy(IP,v[2]);
		else{
			printf("FAILED\nIP address invalid\n");
			exit(1);
		}
//		printf("OK\n");
	}else{
		printf("OK\nInput server IP address:");
		scanf("%s",IP);
		printf("Checking server address(%s)...",IP);
		fflush(stdout);
		sprintf(cmd,"ping -c 1 -W 5 %s >/dev/null 2>/dev/null",IP);
		if(system(cmd)!=0){
			printf("FAILED\n%s is unreachable, please try again later\n",IP);
			exit(1);
		}
	}
	GetSelf(bin);
	printf("OK\nUnpacking install files...");
	fflush(stdout);
	sprintf(cmd,"dd if=%s of=%s skip=%d iflag=skip_bytes >/dev/null 2>/dev/null",bin,TMPFILE,SKIP_BYTE);
	system(cmd);
	sprintf(cmd,"tar xzvf %s -C /tmp >/dev/null 2>/dev/null",TMPFILE);
	system(cmd);
	sprintf(cmd,"rm -f %s >/dev/null 2>/dev/null",TMPFILE);
	system(cmd);

	printf("OK\nInstalling binary files...");
	fflush(stdout);
	system("mkdir -p "INSTALL_DIR);
	if (useubt){
		sprintf(cmd,"/bin/cp %s/ubt/dtdfs %s/ubt/datamgr %s/ubt/cmfs %s >/dev/null 2>/dev/null",TMPDIR,TMPDIR,TMPDIR,INSTALL_DIR);
	}else
		sprintf(cmd,"/bin/cp %s/dtdfs %s/datamgr %s/cmfs %s >/dev/null 2>/dev/null",TMPDIR,TMPDIR,TMPDIR,INSTALL_DIR);
	system(cmd);

	printf("OK\nInstall default container image...");
	fflush(stdout);
	sprintf(cmd,"podman import - cmit <%s/cmit_img.tar >/dev/null 2>/dev/null",TMPDIR);
	system(cmd);

	printf("OK\nConfiguring system...");
	fflush(stdout);

	if (useubt){
		struct stat st;
		if (stat(UBT_CERT,&st)) // not found, install cert
		{
			sprintf(cmd,"/bin/cp %s/cert.pem /usr/local/share/ca-certificates >/dev/null 2>/dev/null",TMPDIR);
			system(cmd);
			system("update-ca-certificates >/dev/null 2>/dev/null");
		}
	}else{
    		sprintf(cmd,"grep `sed -n '2,2p' %s/cert.pem` /etc/pki/tls/certs/ca-bundle.crt >/dev/null 2>/dev/null",TMPDIR);
    		if(system(cmd)){ // not found cert content
        		sprintf(cmd,"cat %s/cert.pem  >>/etc/pki/tls/certs/ca-bundle.crt",TMPDIR);
        		system(cmd);
    		}
    	}

	sprintf(cmd,"echo %s  apisvr  apisvr >>/etc/hosts",IP);
	system(cmd);
	system("rm -rf "TMPDIR);
	printf("OK\nInstall finished, run dtdfs to get more help.\n");
	return 0;	
}
