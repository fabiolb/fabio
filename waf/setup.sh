wget https://raw.githubusercontent.com/jptosso/coraza-waf/v2/master/coraza.conf-recommended -O coraza.conf
git clone https://github.com/coreruleset/coreruleset
mv coreruleset/rules/REQUEST-922-MULTIPART-ATTACK.conf coreruleset/rules/REQUEST-922-MULTIPART-ATTACK.conf.off