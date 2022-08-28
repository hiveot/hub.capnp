# Infrastructure Challenges

The first iterations of HiveOT was intended as a discovery excercise to determine the best course for providing a Hub that is secure, lightweight and easy to scale. These iterations exposed various challenges, especially on the infrastructure side. Below a short summary of these challenges and how to overcome them.

## Security

Security is a core concern. As such the decision is that consumers cannot communicate directly with IoT devices as their security is often lacking.

However, consumers do communicate with the Hub services. Each service has to ensure proper security. If the weakest link fails the Hub is compromised. Good security is tough and when do you know what is good enough? How to keep this level of security every single time a service is written?

Therefore, it is better to solve this once, in middleware, without putting the burden of security on each service.

## Resiliency

Services can fail for various reasons. Bad payload, slow host, lack of memory, and of course, bugs in the code. Once a failure occurs, what should the recovery strategy be?

Ideally the recovery strategy is of little concern to the service. What is handling retries, how to handle failover and how to track failures?

## Scalability

For small systems a single instance of a service is sufficient. But once heavier services are used, and access is given to more consumers, how to scale up and out?

In the IoT space even small computing nodes, like a Raspberry Pi Nano, are powerful enough to run additional tasks, such as transforming a message. How can this free processing power be used?

## Extensibility

Adding services requires quite a bit of understanding of the system as a whole, including security, subscription topics, and more. How to reduce the mental load for adding new features to just what is needed by the service?

## Integration

Integration is another core aspect of the Hub. There is a forest of IoT protocols and the easier integration can take place the better. Can we leverage existing integrations from other solutions?

## Easy to develop

Developers should only have the focus on solving the problem at hand and spend little time on how to integrate it in the whole. As such repetitive tasks such as auth, logging, config, resiliency, security, and others should not be a concern of service developers.

The Hub middleware handles these aspects.

## Open Standards

HiveOT is possible because of open standards like the Linux eco-system. As such it should use open standards where it makes sense.
