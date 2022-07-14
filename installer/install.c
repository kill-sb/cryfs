#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>

#define TMPFILE "/tmp/.dtdfs.tgz"
#define TMPDIR "/tmp/.dtdfs_files"
#define INSTALL_DIR "/usr/local/bin"

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

void Uninstall()
{
	printf("OK\nUninstalling..");
	system("sed -i '/ apisvr /d' /etc/hosts >/dev/null 2>/dev/null");
	system("docker rmi cmit >/dev/null 2>/dev/null");
	system("rm -f /usr/local/bin/dtdfs /usr/local/bin/datamgr /usr/local/bin/cmfs >/dev/null 2>/dev/null");
	printf("OK\nData Defense linux client has been uninstalled from your system\n");
}

int main(int c, char** v)
{
	char bin[4096];
	int docker;
	char IP[1024];
	char cmd[4096];
	int ins=0;

	printf("Checking environment...");
	fflush(stdout);
	ins=Installed();
	// unistall
	if (c==2 && strcmp(v[1],"-u")==0){
		if (ins)
			Uninstall();
		else 
			printf("FAILED\nData Defense is not found in your system\n");
		exit(0);
	}

	// install start
	if (ins){
		printf("FAILED\nData Defense has already been installed\n");
		exit(1);
	}

	docker=system("docker -v 1>/dev/null 2>/dev/null");
	if (docker!=0){
		printf("FAILED\nDocker tools not found, if you are using Centos 8.x, type 'yum install podman-docker' to install it, if you are using other OS, try to use dnf/yum/apt to install docker packges first.\n");
		exit(1);
	}

	if (c>2 && strcmp(v[1],"-svr")==0){
		strcpy(IP,v[2]);
		printf("OK\n");
	}else{
		printf("OK\nInput server IP address:");
		scanf("%s",IP);
	}

	printf("Checking server address(%s)...",IP);
	fflush(stdout);
	sprintf(cmd,"ping -c 1 -W 5 %s >/dev/null 2>/dev/null",IP);
	if(system(cmd)!=0){
		printf("FAILED\n%s is unreachable, please try again later\n",IP);
		exit(1);
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
	sprintf(cmd,"/bin/cp %s/dtdfs %s/datamgr %s/cmfs %s >/dev/null 2>/dev/null",TMPDIR,TMPDIR,TMPDIR,INSTALL_DIR);
	system(cmd);

	printf("OK\nInstall default container image...");
	fflush(stdout);
	sprintf(cmd,"docker import - cmit <%s/cmit_img.tar >/dev/null 2>/dev/null",TMPDIR);
	system(cmd);

	printf("OK\nConfiguring system...");
	fflush(stdout);

    sprintf(cmd,"grep `sed -n '2,2p' %s/cert.pem` /etc/pki/tls/certs/ca-bundle.crt >/dev/null 2>/dev/null",TMPDIR);
    if(system(cmd)){ // not found cert content
        sprintf(cmd,"cat %s/cert.pem  >>/etc/pki/tls/certs/ca-bundle.crt",TMPDIR);
        system(cmd);
    }

	sprintf(cmd,"echo %s  apisvr  apisvr >>/etc/hosts",IP);
	system(cmd);
	system("rm -rf "TMPDIR);
	printf("OK\nInstall finished, run dtdfs to get more help.\n");
	return 0;	
}
