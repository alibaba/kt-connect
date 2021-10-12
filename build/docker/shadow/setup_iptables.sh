#!/bin/bash

redirect_ports=${1}

if [ -n "${redirect_ports}" ]; then
  export KT_REDIRECT_PORTS=${redirect_ports}

  remove_iptables_rules(){
    for pair in $(echo "${KT_REDIRECT_PORTS}" | tr -d ,); do
      echo "remove redirect rule: ${pair}"
      IFS=":" read -r -a ports <<< "${pair}"
      iptables -t nat -D PREROUTING -p tcp --dport ${ports[0]} -j REDIRECT --to-ports ${ports[1]}
    done
  }
  trap 'remove_iptables_rules' SIGTERM

  for pair in $(echo "${redirect_ports}" | tr -d ,); do
    echo "add redirect rule: ${pair}"
    IFS=":" read -r -a ports <<< "${pair}"
    iptables -t nat -A PREROUTING -p tcp --dport "${ports[0]}" -j REDIRECT --to-ports "${ports[1]}"
  done
fi