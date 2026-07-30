/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {describe, expect, it} from 'vitest';
import CustomPlatformTemplateJson from '../../data/application-templates/platform-based/custom.json';
import MCPClientTemplateJson from '../../data/application-templates/technology-based/mcp-client.json';
import {ApplicationCreateFlowStep} from '../../models/application-create-flow';
import type {ApplicationTemplate} from '../../models/application-templates';
import resolveCreationFlow from '../resolveCreationFlow';

const CustomPlatformTemplate = CustomPlatformTemplateJson as ApplicationTemplate;
const MCPClientTemplate = MCPClientTemplateJson as ApplicationTemplate;

describe('resolveCreationFlow', () => {
  it('returns the default user-facing flow (5 steps) when template is null', () => {
    const flow = resolveCreationFlow(null);
    expect(flow.steps).toEqual([
      ApplicationCreateFlowStep.DETAILS,
      ApplicationCreateFlowStep.SECURITY,
      ApplicationCreateFlowStep.DESIGN,
      ApplicationCreateFlowStep.CONFIGURE,
      ApplicationCreateFlowStep.COMPLETE,
    ]);
    expect(flow.previewSteps).toEqual([
      ApplicationCreateFlowStep.DETAILS,
      ApplicationCreateFlowStep.SECURITY,
      ApplicationCreateFlowStep.DESIGN,
    ]);
  });

  it('returns the default user-facing flow when the template has no creationFlow field', () => {
    const flow = resolveCreationFlow({id: 'react', displayName: 'React'});
    expect(flow.steps).toEqual([
      ApplicationCreateFlowStep.DETAILS,
      ApplicationCreateFlowStep.SECURITY,
      ApplicationCreateFlowStep.DESIGN,
      ApplicationCreateFlowStep.CONFIGURE,
      ApplicationCreateFlowStep.COMPLETE,
    ]);
  });

  it('returns the inline creationFlow from the template when present', () => {
    const flow = resolveCreationFlow({
      id: 'backend',
      creationFlow: {
        steps: [ApplicationCreateFlowStep.STACK, ApplicationCreateFlowStep.DETAILS, ApplicationCreateFlowStep.COMPLETE],
        previewSteps: [],
      },
    });
    expect(flow.steps).toEqual([
      ApplicationCreateFlowStep.STACK,
      ApplicationCreateFlowStep.DETAILS,
      ApplicationCreateFlowStep.COMPLETE,
    ]);
  });

  it('returns only DETAILS and COMPLETE steps, with no preview steps, for the custom platform template', () => {
    const flow = resolveCreationFlow(CustomPlatformTemplate);
    expect(flow.steps).toEqual([ApplicationCreateFlowStep.DETAILS, ApplicationCreateFlowStep.COMPLETE]);
    expect(flow.previewSteps).toEqual([]);
  });

  it('returns DETAILS, CLIENT_TYPE, and COMPLETE steps, with no preview steps, for the mcp-client template', () => {
    const flow = resolveCreationFlow(MCPClientTemplate);
    expect(flow.steps).toEqual([
      ApplicationCreateFlowStep.DETAILS,
      ApplicationCreateFlowStep.CLIENT_TYPE,
      ApplicationCreateFlowStep.COMPLETE,
    ]);
    expect(flow.previewSteps).toEqual([]);
  });
});
