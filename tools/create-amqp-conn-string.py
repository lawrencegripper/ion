#!/usr/bin/python3
import urllib.parse
import re
import sys

originalConnString = sys.argv[1]

namespace = re.search( r'sb://[a-zA-Z0-9]*.', originalConnString, re.M|re.I)
if namespace is None:
    print("No namespace match")
    sys.exit(1)
namespace = namespace.group(0).replace("sb://", "")[:-1]

keyName = re.search( r'SharedAccessKeyName=.*;', originalConnString, re.M|re.I)
if keyName is None:
    print("No keyName match")
    sys.exit(1)
keyName = keyName.group(0).replace("SharedAccessKeyName=", "")[:-1]

key = re.search( r'SharedAccessKey=(.*)', originalConnString, re.M|re.I)
if key is None:
    print("No key match")
    sys.exit(1)
key = key.group(0).replace("SharedAccessKey=", "")

escapedKey = urllib.parse.quote_plus(key)
connString = 'amqps://' + keyName + ':' + escapedKey + '@' + namespace + '.servicebus.windows.net/'

print(connString)