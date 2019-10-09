import React, { Component } from 'react';
import { Card, Row, Skeleton, Tag, List, Button, Icon, Modal, Select, Form, Input } from 'antd';
import OwnerReferences from './OwnerReferences/OwnerReferences';
import Terminal from './Terminal/Terminal';
import VirtualService from './VirtualService/VirtualService';
import Log from './Log/Log';
import './index.scss'

const { Meta } = Card;

class AddLocalForm extends Component {

  render() {

    const { getFieldDecorator } = this.props.form;

    return (<Form onSubmit={this.props.handleSubmit}>
      <Form.Item label="Strategy">
        {getFieldDecorator('strategy', {
          rules: [{ required: true, message: 'Please select strategy!' }],
        })(
          <Select defaultValue="headers">
            <Select.Option value="headers">Headers</Select.Option>
          </Select>,
        )}
      </Form.Item>
      <Form.Item label="Match">
        {getFieldDecorator('match', {
          rules: [{ required: true, message: 'Please select match method!' }],
        })(
          <Select defaultValue="extra">
            <Select.Option value="extra">Extra</Select.Option>
            <Select.Option value="prefix">Prefix</Select.Option>
            <Select.Option value="regex">Regex</Select.Option>
          </Select>,
        )}
      </Form.Item>
      <Form.Item label="Value">
        {getFieldDecorator('value', {
          rules: [{ required: true, message: 'Please input match value!' }],
        })(
          <Input
            prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />}
            placeholder="Head Value"
          />,
        )}
      </Form.Item>
    </Form>)
  }
}

class Service extends Component {
  constructor(props) {
    super(props);
    this.state = {
      terminalModelVisible: false,
      logModelVisible: false,
      addLocalRouteVisible: false,
      podName: null,
      containerName: null,
      shell: 'bash',
    }
  }

  componentDidMount() {
    const { match: { params: { name, namespace } } } = this.props;
    this.props.fetchService(namespace, name);
    this.props.fetchVirtualService(namespace, name);
  }

  handleSubmit = () => {
    console.log('handleSubmit');
  }

  render() {
    const { service, endpoint, pods, match: { params: { namespace } } } = this.props;
    const loading = (!service || !endpoint);

    if (loading) {
      return <Skeleton active={loading} />
    }

    const groups = [
      {
        title: 'Pods',
        type: 'cluster',
        pods: pods.filter((pod) => {
          const { metadata: { labels } } = pod;
          return labels['control-by'] !== 'kt';
        })
      }, {
        title: 'Redirect To Wide',
        type: 'system',
        pods: pods.filter((pod) => {
          const { metadata: { labels } } = pod;
          return labels['control-by'] === 'kt';
        })
      }
    ]

    const targetPorts = service.spec.ports.map(port => port.targetPort);

    const AddLocalFormWrapper = Form.create({ name: 'addLocalRouter' })(AddLocalForm);

    return (
      <Row className="container">
        <Card>
          <Meta
            title={
              <div>{
                service.metadata.name}&nbsp;
                <Tag color="#108ee9">{service.spec.type}</Tag>
                <Tag color="#108ee9">{service.spec.clusterIP}</Tag>
                {
                  service.spec.ports.map((port, index) => {
                    return <Tag key={`service.spec.ports.${index}`} color="#108ee9">{`${port.port}:${port.targetPort}`}</Tag>
                  })
                }
              </div>
            }
            description={`${service.metadata.name}.${service.metadata.namespace}.svc.cluster.in`}
          />
          <div>
            <br />
            <p className="command-line">ktctl -namespace {namespace} connect</p>
          </div>
        </Card>
        <br />

        <Modal
          title={`Log:${this.state.pod ? this.state.pod.metadata.name : 'Loading'}`}
          width={860}
          visible={this.state.logModelVisible}
          onOk={() => { this.setState({ logModelVisible: false, containerName: null, podName: null }) }}
          onCancel={() => { this.setState({ logModelVisible: false, containerName: null, podName: null }) }}
          className="log-dialog"
          footer={(this.state.pod && this.state.pod.spec.containers.length > 0) ?
            <div>
              <div className="pull-left">
                Switch Container：
                <Select style={{ width: 120 }} value={this.state.containerName} onChange={(value) => {
                  this.setState({ containerName: value })
                }}>{
                    this.state.pod.spec.containers.map((container, index) => (<Select.Option key={`modal-log-container-${index}`} value={container.name}>{container.name}</Select.Option>))
                  }</Select>
              </div>
            </div> : null
          }
        >
          {this.state.logModelVisible && <Log namespace={namespace} name={this.state.podName} container={this.state.containerName} />}
        </Modal>

        <Modal
          title={`Terminal:${this.state.pod ? this.state.pod.metadata.name : 'Loading'}`}
          width={860}
          visible={this.state.terminalModelVisible}
          onOk={() => { this.setState({ terminalModelVisible: false, containerName: null, podName: null }) }}
          onCancel={() => { this.setState({ terminalModelVisible: false, containerName: null, podName: null }) }}
          className="terminal-dialog"
          footer={(this.state.pod && this.state.pod.spec.containers.length > 0) ?
            <div>
              <div className="pull-left">
                Shell：
                <Select style={{ width: 120 }} defaultValue={this.state.shell} onChange={(shell) => {
                  this.setState({ shell })
                }}>
                  {
                    ['bash', 'sh'].map(item => (
                      <Select.Option key={item} value={item}>{item}</Select.Option>)
                    )
                  }
                </Select>
                &nbsp;
                Switch Container：
                <Select style={{ width: 120 }} value={this.state.containerName} onChange={(value) => {
                  this.setState({ containerName: value })
                }}>{
                    this.state.pod.spec.containers.map((container, index) => (<Select.Option key={`modal-terminal-container-${index}`} value={container.name}>{container.name}</Select.Option>))
                  }</Select>
              </div>
            </div> : null
          }
        >
          {this.state.terminalModelVisible && <Terminal namespace={namespace} name={this.state.podName} container={this.state.containerName} shell={this.state.shell} />}
        </Modal>

        <Modal
          title={`Add Local ${this.state.currentVersion} To Route`}
          visible={this.state.addLocalRouteVisible}
          onOk={() => { this.setState({ addLocalRouteVisible: false }) }}
          onCancel={() => { this.setState({ addLocalRouteVisible: false }) }}
        >
          <AddLocalFormWrapper onSubmit={this.handleSubmit} />
        </Modal>

        <VirtualService virtualservice={this.props.virtualservice} destinationrule={this.props.destinationrule} />

        {groups.map((group, x) => (
          <div key={`pods-group-${x}`}>
            <Card>
              <List
                header={<div>{group.title}</div>}
                dataSource={group.pods}
                itemLayout="horizontal"
                renderItem={(item, index) => {
                  const { metadata: { labels } } = item;
                  const isSystemComponent = labels['control-by'] === 'kt';
                  const isVersion = labels['version'] !== undefined;

                  return (<List.Item
                    key={`pods-${index}`}
                    actions={
                      [
                        <Button shape="circle" onClick={() => {
                          this.setState({ pod: item, podName: item.metadata.name, containerName: item.spec.containers[0].name }, () => {
                            this.setState({ logModelVisible: true })
                          })
                        }}><Icon type="file-text" /></Button>,
                        <Button shape="circle" onClick={() => {
                          this.setState({ pod: item, podName: item.metadata.name, containerName: item.spec.containers[0].name }, () => {
                            this.setState({ terminalModelVisible: true })
                          })
                        }}><Icon type="rocket" /></Button>,
                        (group.type === 'system' && this.props.virtualservice && isVersion) ? <Button shape="circle" onClick={() => {
                          if (!isVersion) {
                            return;
                          }
                          const version = labels['version'];
                          this.setState({ currentVersion: version, addLocalRouteVisible: true })
                        }}><Icon type="swap" /></Button> : null,
                      ]
                    }
                  >
                    <List.Item.Meta
                      title={
                        <div>
                          {item.metadata.name}&nbsp;
                          <Tag>{item.status.phase}</Tag>
                          <Tag>{item.status.podIP}</Tag>
                          {isSystemComponent && <Tag color="#108ee9">{labels['kt-component']}</Tag>}
                          {isVersion && <Tag color="#108ee9">{`version: ${labels['version']}`}</Tag>}
                          {isSystemComponent && <Tag color="#87d068">{`Client: ${labels['remoteAddress']}`}</Tag>}
                        </div>
                      }
                      description={
                        item.metadata.ownerReferences && <OwnerReferences ownerReferences={item.metadata.ownerReferences} namespace={namespace} type={group.type} targetPorts={targetPorts} />
                      }
                    />
                  </List.Item>)
                }}
              />
            </Card>
            <br />
          </div>
        ))}

      </Row>
    );
  }
}

export default Service;
