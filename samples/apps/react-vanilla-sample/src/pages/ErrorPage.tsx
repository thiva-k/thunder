// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Layout from '../components/Layout';

const ErrorPage = ({ errorCode, errorMessage }: { errorCode: string, errorMessage: string }) => {
    return (
        <Layout>
            <Box sx={{ width: '100%' }}>
                <Alert severity="error">
                    <AlertTitle sx={{ mb: 2 }}>Something didn&apos;t go as expected!</AlertTitle>
                    <Typography variant="body1" sx={{ mt: 3 }}>
                        {errorMessage}
                    </Typography>
                    {errorCode !== '' && (
                        <Typography variant="body1" sx={{ mt: 2 }}>
                        Error Code: {errorCode}
                        </Typography>
                    )}
                </Alert>
                <Box sx={{ mt: 4 }}>
                    <Button variant="contained" color="primary" onClick={() => window.location.href = '/'}>
                        Back to Login
                    </Button>
                </Box>
            </Box>
        </Layout>
    );
};

export default ErrorPage;
