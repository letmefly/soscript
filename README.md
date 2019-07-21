# soscript
a little script for project code generation
eg.
// <soscript>
// <default>
let serverAddr = "http://localhost:8888"
// </default>

// <line> if(version == "1.0.1" || mode == "debug") print(<code> let serverAddr = "http://localhost:8888" </code>) </line>
// <line> if(version == "1.0.1" && mode == "release") print(<code> let serverAddr = "http://localhost:8888" </code>) </line>
// <line> if(version == "1.0.1") print(<code> let serverAddr = "http://localhost:8888" </code>) </line>

// <line> switch1 = (version == "1.0.1" && platform == "android") </line>
// <line> switch2 = (version == "1.0.2" && platform == "ios") </line>
// <line> switch3 = (version == "1.0.1" && platform == "wechat") </line>
// <line> if (switch1) => print(<code> let serverAddr = "http://localhost:8888" </code>) </line>
// <line> if (switch2) => print(<code> let serverAddr = "http://localhost:8888" </code>) </line>
// <line> if (switch3) => print(<code> let serverAddr = <var>addr</var> </code>) </line>

// </soscript>


ssc.exe -v def.ss -c wechat_conf.ss -s sss -o oooo
go build . && ssc.exe compile -v def.ss -c wechat_conf.ss -s test.java -o oooo
go build . && ./ssc compile -v def.ss -c wechat_conf.ss -s test.java -o oooo