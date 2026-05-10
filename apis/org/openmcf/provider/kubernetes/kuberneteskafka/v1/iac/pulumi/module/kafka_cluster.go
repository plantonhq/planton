package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/strimzioperator/kubernetes/kafka/v1beta2"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaCluster(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (*v1beta2.Kafka, error) {
	listenersArray := v1beta2.KafkaSpecKafkaListenersArray{}

	listenersArray = append(listenersArray,
		v1beta2.KafkaSpecKafkaListenersArgs{
			Name: pulumi.String(vars.InternalListenerName),
			Port: pulumi.Int(vars.InternalListenerPortNumber),
			Tls:  pulumi.Bool(false),
			Type: pulumi.String("internal"),
			Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
				Type: pulumi.String("scram-sha-512"),
			},
		})

	if locals.KubernetesKafka.Spec.Ingress != nil && locals.KubernetesKafka.Spec.Ingress.Enabled {
		listenersArray = append(listenersArray, getIngressListeners(locals)...)
		//crate new certificate
		certOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err := certmanagerv1.NewCertificate(ctx,
			"kafka-ingress-certificate",
			&certmanagerv1.CertificateArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.KafkaIngressCertName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: certmanagerv1.CertificateSpecArgs{
					DnsNames:   pulumi.ToStringArray(locals.IngressHostnames),
					SecretName: pulumi.String(locals.KafkaIngressCertSecretName),
					IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
						Kind: pulumi.String("ClusterIssuer"),
						Name: pulumi.String(locals.IngressCertClusterIssuerName),
					},
				},
			}, certOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "error creating certificate for bootstrap server ingress")
		}
	}

	// create kafka cluster
	kafkaOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdKafkaCluster, err := v1beta2.NewKafka(ctx,
		"kafka-cluster",
		&v1beta2.KafkaArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KubernetesKafka.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: v1beta2.KafkaSpecArgs{
				EntityOperator: v1beta2.KafkaSpecEntityOperatorArgs{
					//todo: understand resource usage pattern and allocate
					TopicOperator: v1beta2.KafkaSpecEntityOperatorTopicOperatorArgs{},
					UserOperator:  v1beta2.KafkaSpecEntityOperatorUserOperatorArgs{},
				},
				Kafka: v1beta2.KafkaSpecKafkaArgs{
					Authorization: v1beta2.KafkaSpecKafkaAuthorizationArgs{
						SuperUsers: pulumi.StringArray{pulumi.String(locals.AdminUsername)},
						Type:       pulumi.String("simple"),
					},
					Config:    vars.KafkaClusterDefaultConfig,
					Listeners: listenersArray,
					Replicas:  pulumi.Int(locals.KubernetesKafka.Spec.BrokerContainer.Replicas),
					Resources: v1beta2.KafkaSpecKafkaResourcesArgs{
						Limits: pulumi.Map{
							"cpu":    pulumi.String(locals.KubernetesKafka.Spec.BrokerContainer.Resources.Limits.Cpu),
							"memory": pulumi.String(locals.KubernetesKafka.Spec.BrokerContainer.Resources.Limits.Memory),
						},
						Requests: pulumi.Map{
							"cpu":    pulumi.String(locals.KubernetesKafka.Spec.BrokerContainer.Resources.Requests.Cpu),
							"memory": pulumi.String(locals.KubernetesKafka.Spec.BrokerContainer.Resources.Requests.Memory),
						},
					},
					Storage: v1beta2.KafkaSpecKafkaStorageArgs{
						Type: pulumi.String("jbod"),
						Volumes: v1beta2.KafkaSpecKafkaStorageVolumesArray{
							v1beta2.KafkaSpecKafkaStorageVolumesArgs{
								DeleteClaim: pulumi.Bool(false),
								Id:          pulumi.Int(0),
								Size:        pulumi.String(locals.KubernetesKafka.Spec.BrokerContainer.DiskSize),
								Type:        pulumi.String("persistent-claim"),
							},
						},
					},
				},
				Zookeeper: v1beta2.KafkaSpecZookeeperArgs{
					Replicas: pulumi.Int(locals.KubernetesKafka.Spec.BrokerContainer.Replicas),
					Storage: v1beta2.KafkaSpecZookeeperStorageArgs{
						DeleteClaim: pulumi.Bool(false),
						Size:        pulumi.String(vars.ZookeeperDefaultDiskSizeInGb),
						Type:        pulumi.String("persistent-claim"),
					},
				},
			},
		}, kafkaOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka-cluster")
	}
	return createdKafkaCluster, nil
}

// getIngressListeners returns the full list of listeners to be configured on the kafka cluster in order for
// kafka clients outside the kubernetes clusters to be able to talk to the kafka cluster.
// this listeners configuration uses 'load-balancers' approach explained
// at https://strimzi.io/blog/2019/05/13/accessing-kafka-part-4/
func getIngressListeners(locals *Locals) v1beta2.KafkaSpecKafkaListenersArray {
	ingressListenersArray := v1beta2.KafkaSpecKafkaListenersArray{}

	//create two seperate sets of listeners, one for clients with in the same network and
	//another for clients outside the vpc network.

	//depending on the bootstrap server hostname that a client uses, the corresponding
	//broker hostnames are returned to the client.

	//internal load balancer listeners
	ingressListenersArray = append(ingressListenersArray,
		v1beta2.KafkaSpecKafkaListenersArgs{
			Name: pulumi.String(vars.ExternalPrivateListenerName),
			Port: pulumi.Int(vars.ExternalPrivateListenerPortNumber),
			Tls:  pulumi.Bool(true),
			Type: pulumi.String("loadbalancer"),
			Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
				Type: pulumi.String("scram-sha-512"),
			},
			Configuration: &v1beta2.KafkaSpecKafkaListenersConfigurationArgs{
				Bootstrap: &v1beta2.KafkaSpecKafkaListenersConfigurationBootstrapArgs{
					//if a client uses bootstrap server created for clients in the same network(internal), the bootstrap
					//server returns the hostnames of brokers configured in the brokers section below.
					Annotations: pulumi.ToStringMap(map[string]string{
						locals.KafkaIngressPrivateListenerLoadBalancerAnnotationKey: locals.KafkaIngressPrivateListenerLoadBalancerAnnotationValue,
						vars.ExternalDnsHostnameAnnotationKey:                       locals.IngressInternalBootstrapHostname,
					}),
				},
				Brokers: getIngressListenersBrokersArray(locals.IngressInternalBrokerHostnames,
					locals.KafkaIngressPrivateListenerLoadBalancerAnnotationKey,
					locals.KafkaIngressPrivateListenerLoadBalancerAnnotationValue),
				BrokerCertChainAndKey: &v1beta2.KafkaSpecKafkaListenersConfigurationBrokerCertChainAndKeyArgs{
					Certificate: pulumi.String("tls.crt"),
					Key:         pulumi.String("tls.key"),
					SecretName:  pulumi.String(locals.KafkaIngressCertSecretName),
				},
			},
		})

	//public load balancer listeners
	ingressListenersArray = append(ingressListenersArray, v1beta2.KafkaSpecKafkaListenersArgs{
		Name: pulumi.String(vars.ExternalPublicListenerName),
		Port: pulumi.Int(vars.ExternalPublicListenerPortNumber),
		Tls:  pulumi.Bool(true),
		Type: pulumi.String("loadbalancer"),
		Authentication: &v1beta2.KafkaSpecKafkaListenersAuthenticationArgs{
			Type: pulumi.String("scram-sha-512"),
		},
		Configuration: &v1beta2.KafkaSpecKafkaListenersConfigurationArgs{
			Bootstrap: &v1beta2.KafkaSpecKafkaListenersConfigurationBootstrapArgs{
				//if a client uses bootstrap server created for clients 'outside' the network, the bootstrap
				//server returns the hostnames of brokers configured in the brokers section below.
				Annotations: pulumi.ToStringMap(map[string]string{
					locals.KafkaIngressPublicListenerLoadBalancerAnnotationKey: locals.KafkaIngressPublicListenerLoadBalancerAnnotationValue,
					vars.ExternalDnsHostnameAnnotationKey:                      locals.IngressExternalBootstrapHostname,
				}),
			},
			Brokers: getIngressListenersBrokersArray(locals.IngressExternalBrokerHostnames,
				locals.KafkaIngressPublicListenerLoadBalancerAnnotationKey,
				locals.KafkaIngressPublicListenerLoadBalancerAnnotationValue),
			BrokerCertChainAndKey: &v1beta2.KafkaSpecKafkaListenersConfigurationBrokerCertChainAndKeyArgs{
				Certificate: pulumi.String("tls.crt"),
				Key:         pulumi.String("tls.key"),
				SecretName:  pulumi.String(locals.KafkaIngressCertSecretName),
			},
		},
	})
	return ingressListenersArray
}

func getIngressListenersBrokersArray(hostnames []string, loadBalancerAnnotationKey,
	loadBalancerAnnotationValue string) v1beta2.KafkaSpecKafkaListenersConfigurationBrokersArray {
	resp := make([]v1beta2.KafkaSpecKafkaListenersConfigurationBrokersInput, len(hostnames))
	for i, hostName := range hostnames {
		resp[i] = v1beta2.KafkaSpecKafkaListenersConfigurationBrokersArgs{
			Broker:         pulumi.Int(i),
			AdvertisedHost: pulumi.String(hostName),
			AdvertisedPort: pulumi.Int(vars.ExternalPublicListenerPortNumber),
			Annotations: pulumi.ToStringMap(map[string]string{
				loadBalancerAnnotationKey:             loadBalancerAnnotationValue,
				vars.ExternalDnsHostnameAnnotationKey: hostName,
			}),
		}
	}
	return resp
}
