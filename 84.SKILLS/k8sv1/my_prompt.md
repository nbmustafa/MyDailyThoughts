Act as a principal claude architect and Ai Engineer to write a comprehensive k8s diognise tool with golang for quickly diognisting any kubermetes issues.
Details: 
- the tool expect cluster name, then call a golang script for overall checks across entire cluster for any issues and report it out the result of findings in nice and coloured format.

- if you peovidr cluster name and namespace in the command as parameter it should scope the investigationon that specific namespace.

- the investigation should include pod related issues, storage,pv, pvc related issue, node related issuss, affinity, node selection, service endpoint, label related issues or netpolicy, and calico netpolicy related issue and any other possible area to be looked to make sure the tool covers all asprct and make it a complete and comprehensive diognistic tool

- namespace event to be checked carefully

Please follow the best practice of software development and consider that in repository structure of the golang for this tool