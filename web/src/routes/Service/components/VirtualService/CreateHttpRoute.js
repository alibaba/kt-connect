import React, { Component } from "react";
import { Button, Modal, Steps, Transfer, Radio } from 'antd';

const { Step } = Steps;

const mockData = [];
for (let i = 0; i < 10; i++) {
  mockData.push({
    key: i.toString(),
    title: `content${i + 1}`,
    description: `description of content${i + 1}`,
    disabled: i % 3 < 1,
  });
}

const oriTargetKeys = mockData.filter(item => +item.key % 3 > 1).map(item => item.key);

export default class CreateHttpRoute extends Component {

  constructor(props) {
    super(props);
    console.log(props);
    this.state = {
      current: 0,
      targetKeys: oriTargetKeys,
      selectedKeys: [],
      template: 'system',
    }
  }

  handleChange = (nextTargetKeys, direction, moveKeys) => {
    this.setState({ targetKeys: nextTargetKeys });
  };

  handleSelectChange = (sourceSelectedKeys, targetSelectedKeys) => {
    this.setState({ selectedKeys: [...sourceSelectedKeys, ...targetSelectedKeys] });
  };

  render() {
    const { current = 0, targetKeys, selectedKeys } = this.state;
    const steps = [
      {
        name: 'Destination',
        render: () => {
          return (
            <div>
              <Transfer
                dataSource={mockData}
                titles={['Source', 'Target']}
                targetKeys={targetKeys}
                selectedKeys={selectedKeys}
                onChange={this.handleChange}
                onSelectChange={this.handleSelectChange}
                render={item => item.title}
              />
            </div>
          )
        },
        buttons: [
          <Button key="step1-pre" onClick={() => { this.setState({ current: this.state.current + 1 }) }}>Next</Button>
        ]
      },
      {
        name: 'policy',
        render: () => {
          return (
            <div>
              <div>
                <Radio.Group onChange={(e) => {
                  this.setState({ template: e.target.value });
                }} value={this.state.template}>
                  <Radio value='system'>System</Radio>
                  <Radio value='custom'>Custom</Radio>
                </Radio.Group>
              </div>
            </div>
          )
        },
        buttons: [
          <Button key="step2-pre" onClick={() => { this.setState({ current: this.state.current - 1 }) }}>Pre</Button>,
          <Button key="step2-next" onClick={() => { this.setState({ current: this.state.current + 1 }) }}>Next</Button>
        ]
      },
      {
        name: 'preview',
        render: () => {
          return '3'
        },
        buttons: [
          <Button key="step3-pre" onClick={() => { this.setState({ current: this.state.current - 1 }) }}>Pre</Button>,
          <Button key="step3-next" onClick={() => { this.setState({ current: this.state.current + 1 }) }}>Next</Button>
        ]
      },
      {
        name: 'done',
        render: () => {
          return '4'
        },
        buttons: [
          <Button key="step4-pre" onClick={() => { this.setState({ current: this.state.current - 1 }) }}>Pre</Button>,
          <Button key="step3-finished">Finished</Button>
        ],
      }
    ];

    const currentStep = steps[current];

    return (
      <Modal
        title="Add Route"
        visible={this.props.visible}
        onOk={() => {
          this.props.onOk();
          this.setState({ current: 0 });
        }}
        onCancel={() => {
          this.props.onCancel();
          this.setState({ current: 0 });
        }}
        width={860}
        footer={currentStep.buttons}
      >
        <Steps current={current}>
          {
            steps.map((step, index) => {
              return <Step key={`step-${index}`} title={step.name} />
            })
          }
        </Steps>
        <div className="steps-content">{currentStep.render()}</div>
        <div className="steps-footer">
        </div>
      </Modal>
    );
  }
}