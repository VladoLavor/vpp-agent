*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../../variables/${VARIABLES}_variables.robot

Resource     ../../../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${FINAL_SLEEP}=        3s
${SYNC_SLEEP}=         10s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Start SFC Controller Container With Own Config    basic.conf
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    Start SFC Controller Container With Own Config    basic.conf
    Sleep    ${SYNC_SLEEP}

Check Memif Interface On VPP2
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=02:02:02:02:02:02  role=slave  ipv4=10.0.0.1/24  connected=1  enabled=1

Check Memif Interface On VPP3
    vat_term: Check Memif Interface State     agent_vpp_3  vpp3_memif1  role=slave  ipv4=10.0.0.10/24  connected=1  enabled=1 

Show Interfaces And Other Objects After Config
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    vpp_term: Show Interfaces    agent_vpp_3
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_2_term    show int addr
    Write To Machine    agent_vpp_3_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_2_term    show h
    Write To Machine    agent_vpp_3_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_2_term    show br
    Write To Machine    agent_vpp_3_term    show br
    Write To Machine    agent_vpp_1_term    show br 1 detail
    Write To Machine    agent_vpp_2_term    show br 1 detail
    Write To Machine    agent_vpp_3_term    show br 1 detail
    Write To Machine    agent_vpp_1_term    show vxlan tunnel
    Write To Machine    agent_vpp_2_term    show vxlan tunnel
    Write To Machine    agent_vpp_3_term    show vxlan tunnel
    Write To Machine    agent_vpp_1_term    show err
    Write To Machine    agent_vpp_2_term    show err
    Write To Machine    agent_vpp_3_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    vat_term: Interfaces Dump    agent_vpp_2
    vat_term: Interfaces Dump    agent_vpp_3
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a
    Execute In Container    agent_1    ip a

Check Ping Agnet2 -> Agent3
    vpp_term: Check Ping    agent_vpp_2    10.0.0.10

Check Ping Agnet3 -> Agent2
    vpp_term: Check Ping    agent_vpp_3    10.0.0.1

Done
    [Tags]    debug
    No Operation

Final Sleep For Manual Checking
    [Tags]    debug
    Sleep   ${FINAL_SLEEP}

*** Keywords ***
Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown
