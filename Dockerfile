FROM rabbitmq:3.6.6-management

ADD plugins/rabbitmq_delayed_message_exchange-0.0.1.ez /usr/lib/rabbitmq/lib/rabbitmq_server-3.6.6/plugins/

RUN rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange
