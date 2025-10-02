#/bin/bash
(
    cd /tmp
    rm -f dns_ch.txt ch.txt

    dig -y hmac-sha512:tsig-zonedata-ch-public-21-01:stZwEGApYumtXkh73qMLPqfbIDozWKZLkqRvcjKSpRnsor6A6MxixRL6C2HeSVBQNfMW4wer+qjS0ZSfiWiJ3Q== @zonedata.switch.ch +noall +answer +noidnout +onesoa AXFR ch. > dns_ch.txt
    grep -v "RRSIG" dns_ch.txt | awk '{print $1}' | sed 's/\.$//' | uniq > ch.txt
)

./split /tmp/ch.txt
