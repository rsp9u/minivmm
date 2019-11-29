#!/bin/bash -e

#
# NOTE: This works well with OpenStack installed via kolla.
#

usage() {
  echo "$0 <instance name> <instance id> <cpu> <memory> <disk>"
}

if [ $# -lt 5 ]; then
  usage
  exit 1
fi

DIR=$PWD/vms/$1
mkdir -p $DIR


#### Copy and merge the disk files
snapshot=/var/lib/docker/volumes/nova_compute/_data/instances/$2/disk
backing=$(qemu-img info $snapshot | grep backing | cut -d: -f2 | tr -d " " | sed -e 's,/var/lib/nova/instances,/var/lib/docker/volumes/nova_compute/_data/instances,')
backing_dst=$DIR/${1}-backing.qcow2
image_dst=$DIR/${1}.qcow2

ls $(qemu-img info $snapshot | grep backing | cut -d: -f2 | tr -d " ") > /dev/null 2>&1
if [ $? -ne 0 ]; then
  # to restore original backing file path for rebasing
  mkdir -p /var/lib/nova/instances/
  ln -s /var/lib/docker/volumes/nova_compute/_data/instances/_base /var/lib/nova/instances/_base
fi

echo "copying backing.."
cp $backing $backing_dst

echo "rebasing.."
qemu-img rebase -b $backing_dst $snapshot
echo "committing.. (this would take a long time)"
qemu-img commit $snapshot
echo "reconverting to qcow2.. (this would take a long time)"
qemu-img convert -f raw -O qcow2 $backing_dst $image_dst
rm -f $backing_dst


#### Create cloud-init userdata ISO
echo "creating cloud-init iso.."
iso=$DIR/cloud-init.iso
ud=/tmp/user-data
md=/tmp/meta-data
touch $ud $md
genisoimage -output $iso -volid cidata -joliet -rock -input-charset utf-8 $ud $md > /dev/null 2>&1
rm -f $ud $md


#### Create a metadata file
echo "creating metadata file.."
python -c "
import os,json,binascii

mac = ':'.join([binascii.hexlify(os.urandom(1)) for i in range(3)])
data = {
    'name': '$1',
    'status': 'stopping',
    'owner': '',
    'image': '',
    'volume': '/opt/infra/minivmm/vms/$1/${1}.qcow2',
    'mac_address': '52:54:00:' + mac,
    'ip_address': '',
    'cpu': '$3',
    'memory': '$4',
    'disk': '$5',
    'tag': '',
    'vnc_password': '',
    'vnc_port': '',
    'user_data': '',
    'cloud_init_iso': '/opt/infra/minivmm/vms/$1/cloud-init.iso',
}

with open('$DIR/metadata.json', 'w') as f:
    json.dump(data, f)
"
