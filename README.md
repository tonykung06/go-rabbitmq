###Commands
- `http://www.rabbitmq.com/man/rabbitmqctl.1.man.html`
- `http://www.rabbitmq.com/man/rabbitmq-plugins.1.man.html`
- `rabbitmq-server`
- `rabbitmqctl status`
- `rabbitmqctl list_queues`
- `rabbitmqctl cluster_status`
- `rabbitmqctl stop`
- `rabbitmq-service start`
- `rabbitmq-plugins list`
- `rabbitmq-plugins disable rabbitmq_management`
- `rabbitmq-plugins enable rabbitmq_management`
- default rabbitmq management console log credentials: `guest:guest`
###Run the program
- `go run ./sensors/sensor.go --help`
- `start go run ./coordinator/exec/main.go`

###Rabbitmq server runs on port 5672 by default, and the Web Management UI works on 15672, the default login credential is guest:guest

###rabbitmqctl
- stop
- reset
- stop_app
- start_app
- various user management

###rabbitmq-service
- stop
- start
- install
