// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useGetApplications} from '@thunderid/configure-applications';
import type {BasicApplication} from '@thunderid/configure-applications';
import {useGetUsers} from '@thunderid/configure-users';
import {useDataGridLocaleText} from '@thunderid/hooks';
import type {User} from '@thunderid/types';
import {getErrorMessage} from '@thunderid/utils';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Alert,
  DataGrid,
  Avatar,
  Chip,
  Typography,
  Tabs,
  Tab,
  useTheme,
} from '@wso2/oxygen-ui';
import {AppWindow, Bot, User as UserIcon, Users} from '@wso2/oxygen-ui-icons-react';
import {useState, useMemo, useCallback, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import useGetAgents from '../../../../agents/api/useGetAgents';
import type {BasicAgent} from '../../../../agents/models/agent';
import useGetGroups from '../../../api/useGetGroups';
import type {GroupBasic, Member} from '../../../models/group';

interface AddMemberDialogProps {
  open: boolean;
  onClose: () => void;
  onAdd: (members: Member[]) => void;
  /** Group being edited, excluded from the groups tab so it cannot be made a member of itself. */
  excludeGroupId?: string;
  /** Inline error shown in the dialog when the last add attempt failed. */
  error?: string | null;
  /** Called when the tab or a selection changes, so the parent can clear a stale error. */
  onErrorDismiss?: () => void;
  /** Whether the add mutation is in flight, so the confirm button can show progress. */
  isSubmitting?: boolean;
}

/**
 * Dialog for searching and adding user, app, agent, or group members to a group.
 */
export default function AddMemberDialog({
  open,
  onClose,
  onAdd,
  excludeGroupId = undefined,
  error = null,
  onErrorDismiss = undefined,
  isSubmitting = false,
}: AddMemberDialogProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();

  const [activeTab, setActiveTab] = useState(0);
  const [userSelectionModel, setUserSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [appSelectionModel, setAppSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [agentSelectionModel, setAgentSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [groupSelectionModel, setGroupSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [userPaginationModel, setUserPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [appPaginationModel, setAppPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [agentPaginationModel, setAgentPaginationModel] = useState<DataGrid.GridPaginationModel>({
    pageSize: 10,
    page: 0,
  });
  const [groupPaginationModel, setGroupPaginationModel] = useState<DataGrid.GridPaginationModel>({
    pageSize: 10,
    page: 0,
  });

  const usersParams = useMemo(
    () => ({
      limit: userPaginationModel.pageSize,
      offset: userPaginationModel.page * userPaginationModel.pageSize,
    }),
    [userPaginationModel],
  );
  const appsParams = useMemo(
    () => ({
      limit: appPaginationModel.pageSize,
      offset: appPaginationModel.page * appPaginationModel.pageSize,
    }),
    [appPaginationModel],
  );
  const agentsParams = useMemo(
    () => ({
      limit: agentPaginationModel.pageSize,
      offset: agentPaginationModel.page * agentPaginationModel.pageSize,
    }),
    [agentPaginationModel],
  );
  const groupsParams = useMemo(
    () => ({
      limit: groupPaginationModel.pageSize,
      offset: groupPaginationModel.page * groupPaginationModel.pageSize,
    }),
    [groupPaginationModel],
  );
  const {data: usersData, isLoading: usersLoading, error: usersError} = useGetUsers(usersParams);
  const {data: appsData, isLoading: appsLoading, error: appsError} = useGetApplications(appsParams);
  const {data: agentsData, isLoading: agentsLoading, error: agentsError} = useGetAgents(agentsParams);
  const {data: groupsData, isLoading: groupsLoading, error: groupsError} = useGetGroups(groupsParams);

  const users: User[] = useMemo(() => usersData?.users ?? [], [usersData]);
  const applications: BasicApplication[] = useMemo(() => appsData?.applications ?? [], [appsData]);
  const agents: BasicAgent[] = useMemo(() => agentsData?.agents ?? [], [agentsData]);
  // The listing is server-paginated, so the edited group is dropped from the current page and the
  // total is adjusted by one rather than being filtered out of the query itself.
  const groups: GroupBasic[] = useMemo(
    () => (groupsData?.groups ?? []).filter((group) => group.id !== excludeGroupId),
    [groupsData, excludeGroupId],
  );
  const groupsRowCount: number = useMemo(() => {
    const total = groupsData?.totalResults ?? 0;
    return excludeGroupId ? Math.max(total - 1, 0) : total;
  }, [groupsData, excludeGroupId]);

  const userColumns: DataGrid.GridColDef<User>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {
                  backgroundColor: theme.vars?.palette.grey[900],
                }),
              }}
            >
              <UserIcon size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'display',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<User>): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              height: '100%',
              overflow: 'hidden',
            }}
          >
            <Typography variant="body2" noWrap>
              {params.row.display ?? params.row.id}
            </Typography>
            <Typography
              variant="caption"
              color="text.secondary"
              noWrap
              sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
            >
              {params.row.id}
            </Typography>
          </Box>
        ),
      },
      {
        field: 'type',
        headerName: t('groups:addMember.columns.userType'),
        width: 150,
        renderCell: (params: DataGrid.GridRenderCellParams<User>): JSX.Element => (
          <Chip label={params.row.type} size="small" variant="outlined" sx={{textTransform: 'capitalize'}} />
        ),
      },
    ],
    [theme, t],
  );

  const agentColumns: DataGrid.GridColDef<BasicAgent>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {
                  backgroundColor: theme.vars?.palette.grey[900],
                }),
              }}
            >
              <Bot size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme, t],
  );

  const groupColumns: DataGrid.GridColDef<GroupBasic>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {
                  backgroundColor: theme.vars?.palette.grey[900],
                }),
              }}
            >
              <Users size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme, t],
  );

  const appColumns: DataGrid.GridColDef<BasicApplication>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {
                  backgroundColor: theme.vars?.palette.grey[900],
                }),
              }}
            >
              <AppWindow size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme, t],
  );

  const handleAdd = useCallback(() => {
    const newMembers: Member[] = [
      ...[...userSelectionModel.ids].map((id) => ({id: String(id), type: 'user' as const})),
      ...[...appSelectionModel.ids].map((id) => ({id: String(id), type: 'app' as const})),
      ...[...agentSelectionModel.ids].map((id) => ({id: String(id), type: 'agent' as const})),
      ...[...groupSelectionModel.ids].map((id) => ({id: String(id), type: 'group' as const})),
    ];
    // Selections are deliberately left as-is here: the dialog unmounts on a successful add (the
    // parent closes it), and on failure the user needs their selection intact to retry.
    onAdd(newMembers);
  }, [userSelectionModel, appSelectionModel, agentSelectionModel, groupSelectionModel, onAdd]);

  const handleClose = (): void => {
    // Also reached via Escape and backdrop clicks, so guard the in-flight case here rather than
    // only disabling Cancel: closing mid-request would discard the selection needed to retry.
    if (isSubmitting) return;

    setUserSelectionModel({type: 'include', ids: new Set()});
    setAppSelectionModel({type: 'include', ids: new Set()});
    setAgentSelectionModel({type: 'include', ids: new Set()});
    setGroupSelectionModel({type: 'include', ids: new Set()});
    onClose();
  };

  const totalSelected =
    userSelectionModel.ids.size +
    appSelectionModel.ids.size +
    agentSelectionModel.ids.size +
    groupSelectionModel.ids.size;

  const handleTabChange = (_event: SyntheticEvent, tab: number): void => {
    setActiveTab(tab);
    onErrorDismiss?.();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('groups:addMember.title')}</DialogTitle>
      <DialogContent>
        <Tabs value={activeTab} onChange={handleTabChange} sx={{mb: 2}}>
          <Tab icon={<UserIcon size={16} />} iconPosition="start" label={t('groups:addMember.tabs.users')} />
          <Tab icon={<AppWindow size={16} />} iconPosition="start" label={t('groups:addMember.tabs.apps')} />
          <Tab icon={<Bot size={16} />} iconPosition="start" label={t('groups:addMember.tabs.agents', 'Agents')} />
          <Tab icon={<Users size={16} />} iconPosition="start" label={t('groups:addMember.tabs.groups', 'Groups')} />
        </Tabs>

        {activeTab === 0 && (
          <>
            {usersError && !usersLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {getErrorMessage(
                  usersError,
                  t,
                  'groups:addMember.fetchError',
                  'Failed to load users. Please try again.',
                )}
              </Alert>
            )}
            {!usersError && users.length === 0 && !usersLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResults')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={users}
                columns={userColumns}
                loading={usersLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={userSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setUserSelectionModel(newSelection);
                  onErrorDismiss?.();
                }}
                paginationMode="server"
                rowCount={usersData?.totalResults ?? 0}
                paginationModel={userPaginationModel}
                onPaginationModelChange={setUserPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 1 && (
          <>
            {appsError && !appsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {getErrorMessage(
                  appsError,
                  t,
                  'groups:addMember.fetchAppsError',
                  'Failed to load apps. Please try again.',
                )}
              </Alert>
            )}
            {!appsError && applications.length === 0 && !appsLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResultsApps')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={applications}
                columns={appColumns}
                loading={appsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={appSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setAppSelectionModel(newSelection);
                  onErrorDismiss?.();
                }}
                paginationMode="server"
                rowCount={appsData?.totalResults ?? 0}
                paginationModel={appPaginationModel}
                onPaginationModelChange={setAppPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 2 && (
          <>
            {agentsError && !agentsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {getErrorMessage(
                  agentsError,
                  t,
                  'groups:addMember.fetchAgentsError',
                  'Failed to load agents. Please try again.',
                )}
              </Alert>
            )}
            {!agentsError && agents.length === 0 && !agentsLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResultsAgents', 'No agents found')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={agents}
                columns={agentColumns}
                loading={agentsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={agentSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setAgentSelectionModel(newSelection);
                  onErrorDismiss?.();
                }}
                paginationMode="server"
                rowCount={agentsData?.totalResults ?? 0}
                paginationModel={agentPaginationModel}
                onPaginationModelChange={setAgentPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 3 && (
          <>
            {groupsError && !groupsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {getErrorMessage(
                  groupsError,
                  t,
                  'groups:addMember.fetchGroupsError',
                  'Failed to load groups. Please try again.',
                )}
              </Alert>
            )}
            {!groupsError && groups.length === 0 && !groupsLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResultsGroups', 'No groups found')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={groups}
                columns={groupColumns}
                loading={groupsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={groupSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setGroupSelectionModel(newSelection);
                  onErrorDismiss?.();
                }}
                paginationMode="server"
                rowCount={groupsRowCount}
                paginationModel={groupPaginationModel}
                onPaginationModelChange={setGroupPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}
      </DialogContent>
      {error && (
        <Box sx={{px: 3, pt: 2}}>
          <Alert severity="error">{error}</Alert>
        </Box>
      )}
      <DialogActions>
        <Button onClick={handleClose} disabled={isSubmitting}>
          {t('common:actions.cancel', 'Cancel')}
        </Button>
        <Button variant="contained" onClick={handleAdd} disabled={totalSelected === 0 || isSubmitting}>
          {t('groups:addMember.add', 'Add Selected')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
