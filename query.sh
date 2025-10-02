#/bin/bash

dig -y hmac-sha512:tsig-zonedata-ch-public-21-01:stZwEGApYumtXkh73qMLPqfbIDozWKZLkqRvcjKSpRnsor6A6MxixRL6C2HeSVBQNfMW4wer+qjS0ZSfiWiJ3Q== @zonedata.switch.ch +noall +answer +noidnout +onesoa AXFR ch. > _ch.txt
grep -v "RRSIG" _ch.txt | awk '{print $1}' | sed 's/\.$//' | uniq > ch.txt
