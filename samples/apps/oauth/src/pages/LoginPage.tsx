/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import Alert from '@mui/material/Alert';
import InputAdornment from '@mui/material/InputAdornment';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Checkbox from '@mui/material/Checkbox';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import OutlinedInput from '@mui/material/OutlinedInput';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';
import GoogleIcon from '@mui/icons-material/Google';
import GitHubIcon from '@mui/icons-material/GitHub';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import { useEffect, useRef, useState, useCallback } from 'react';
import Layout from '../components/Layout';
import ConnectionErrorModal from '../components/ConnectionErrorModal';
import { 
    NativeAuthSubmitType, 
    initiateNativeAuthFlow,
    initiateNativeAuthFlowWithData,
    submitNativeAuth, 
    submitAuthDecision 
} from '../services/authService';
import useAuth from '../hooks/useAuth';

// Define the interface for the authentication input
interface AuthInput {
    name: string;
    type: string;
    required: boolean;
}

interface ActionPrompt {
    type: string;
    id: string;
}

// Define the interface for the authentication response
interface AuthResponse {
    flowStatus?: string;
    assertion?: string;
    failureReason?: string;
    type?: string;
    data?: {
        actions?: ActionPrompt[];
        inputs?: AuthInput[];
        redirectURL?: string;
        additionalData?: {
            idpName?: string;
        };
    };
    flowId?: string;
}

// Define the interface for the submission error
interface SubmissionError {
    code?: string;
    message?: string;
    description?: string;
}

/**
 * LoginPage component renders the login page with dynamic options based on the server response.
 */
const LoginPage = () => {

    const START_INIT_KEY = 'startInit';
    const FLOW_ID_KEY = 'flowId';

    const isComponentReMount = useRef(false);
    const { setToken, clearToken } = useAuth();

    const [showRememberMe] = useState<boolean>(false);
    const [showForgotPassword] = useState<boolean>(false);
    const [error, setError] = useState<boolean>(false);
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [connectionError, setConnectionError] = useState<boolean>(false);

    const [loading, setLoading] = useState<boolean>(true);
    const [retryCount, setRetryCount] = useState<number>(0);
    const [flowId, setFlowId] = useState<string>(sessionStorage.getItem(FLOW_ID_KEY) || '');
    const [startInit] = useState<boolean>(JSON.parse(sessionStorage.getItem(START_INIT_KEY) || 'true'));

    // Unified form data state
    const [formData, setFormData] = useState<Record<string, string>>({});
    const [inputs, setInputs] = useState<AuthInput[]>([]);
    const [showPassword, setShowPassword] = useState(false);

    // Add new state variable to track redirection URL
    const [redirectURL, setRedirectURL] = useState<string | null>(null);
    const [socialIdpName, setSocialIdpName] = useState<string>('');

    // Add new state variables to track auth flow
    const [needsDecision, setNeedsDecision] = useState<boolean>(false);
    const [availableActions, setAvailableActions] = useState<ActionPrompt[]>([]);
    const [selectedAction, setSelectedAction] = useState<string | null>(null);
    
    // Add state to track signup mode
    const [isSignupMode, setIsSignupMode] = useState<boolean>(false);
    const [regOnlySuccess, setRegOnlySuccess] = useState<boolean>(false);
    const [promptRegistration, setPromptRegistration] = useState<boolean>(false);
    
    const GradientCircularProgress = () => {
        return (
          <>
            <svg width={0} height={0}>
              <defs>
                <linearGradient id="my_gradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" stopColor="#fc4700" />
                  <stop offset="100%" stopColor="#f87643" />
                </linearGradient>
              </defs>
            </svg>
            <CircularProgress sx={{ 'svg circle': { stroke: 'url(#my_gradient)' } }} />
          </>
        );
    }

    // OTP input handling
    const otpLength = 6;
    const [otpDigits, setOtpDigits] = useState(Array(otpLength).fill(''));
    const otpInputRefs = useRef<(HTMLInputElement | null)[]>([]);
    
    useEffect(() => {
        // Initialize refs array
        otpInputRefs.current = otpInputRefs.current.slice(0, otpDigits.length);
    }, [otpDigits.length]);
    
    const handleOTPDigitChange = (index: number, value: string) => {
        // Only allow a single digit
        if (value.length > 1) {
            value = value.slice(0, 1);
        }
        
        // Update the digit at the specified index
        const newOtpDigits = [...otpDigits];
        newOtpDigits[index] = value;
        setOtpDigits(newOtpDigits);
        
        // Combine digits for the form data
        const combinedOtp = newOtpDigits.join('');
        setFormData(prev => ({ ...prev, otp: combinedOtp }));
        
        // Auto-advance to next input if a digit was entered
        if (value && index < otpDigits.length - 1) {
            otpInputRefs.current[index + 1]?.focus();
        }
    };
    
    const handleOTPKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
        // Handle backspace to go to previous input
        if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
            otpInputRefs.current[index - 1]?.focus();
        }
    };
    
    const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
        e.preventDefault();
        const pastedData = e.clipboardData.getData('text');
        
        // Only process if the pasted content looks like a valid OTP
        if (pastedData.match(/^\d+$/) && pastedData.length <= otpDigits.length) {
            const newOtpDigits = [...otpDigits];
            
            // Fill in the digits from the pasted content
            for (let i = 0; i < pastedData.length; i++) {
                newOtpDigits[i] = pastedData[i];
            }
            
            setOtpDigits(newOtpDigits);
            setFormData(prev => ({ ...prev, otp: newOtpDigits.join('') }));
            
            // Focus the next empty digit or the last digit if all filled
            const nextEmptyIndex = pastedData.length < otpDigits.length ? pastedData.length : otpDigits.length - 1;
            otpInputRefs.current[nextEmptyIndex]?.focus();
        }
    };

    // Effect to focus on the first OTP input when available.
    useEffect(() => {
        const hasOTPInput = inputs.some(input => input.type === "otp" || input.name === "otp");
        
        if (hasOTPInput && otpInputRefs.current && otpInputRefs.current.length > 0) {
            setTimeout(() => {
                if (otpInputRefs.current[0]) {
                    otpInputRefs.current[0].focus();
                }
            }, 100);
        }
    }, [inputs]);

    // Single handler for all input changes
    const handleInputChange = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const { name, value } = event.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleTogglePasswordVisibility = () => {
        setShowPassword((prev) => !prev);
    };

    // To prevent focus loss of show/hide password toggle button
    const handleMouseDownPassword = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.preventDefault(); // Prevent focus loss
    };

    const handleSocialLoginClick = useCallback((redirectURL: string) => {
        setLoading(true);
        sessionStorage.setItem(FLOW_ID_KEY, flowId);
        sessionStorage.setItem(START_INIT_KEY, "false");
        window.location.href = redirectURL;
    }, [flowId]);

    // Process authentication response
    const processAuthResponse = useCallback((data: AuthResponse, selectedAction?: string) => {
        const isCameFromDecision = needsDecision;
        const isMobileLogin = selectedAction && selectedAction.includes('mobile');

        setFlowId(data.flowId || '');
        if (data.flowStatus && data.flowStatus == 'ERROR') {
            if (isMobileLogin && data?.failureReason && data.failureReason.includes("User not found")) {
                console.log("User not found, prompting registration");
                setPromptRegistration(true);
                setError(false);
                setErrorMessage('');
                setLoading(false);
                return;
            }

            const defaultMessage = isSignupMode 
                ? 'Registration failed. Please check your information.' 
                : 'Login failed. Please check your credentials.';
            setError(true);
            setErrorMessage(data.failureReason || defaultMessage);
            setLoading(false);
            return;
        }

        // Clear previous state
        clearToken();
        setError(false);
        setConnectionError(false);
        setNeedsDecision(false);
        setFormData({});
        setAvailableActions([]);
        setInputs([]);
        setRedirectURL(null);
        setSocialIdpName('');
        setRegOnlySuccess(false);

        if (data.flowStatus && data.flowStatus === 'COMPLETE') {
            setError(false);
            if (data.assertion) {
                setToken(data.assertion);
            } else {
                setRegOnlySuccess(true);
            }
        } else if (data.type === "VIEW") {
            // Handle the VIEW response
            if (data.data?.actions) {
                // This is a decision screen
                setNeedsDecision(true);
                setAvailableActions(data.data.actions);
            } else if (data.data?.inputs) {
                // This is an input prompt
                setNeedsDecision(false);
                data.data.inputs.forEach((input: AuthInput) => {
                    setInputs(prev => [...prev, input]);
                });
            }
        } else if (data.type === "REDIRECTION") {
            // Handle redirection for social logins
            const url = data.data?.redirectURL;
            const idpName = data.data?.additionalData?.idpName || 'Social Login';

            if (isCameFromDecision) {
                // If this is a decision screen, handle the social login click
                handleSocialLoginClick(url || '');
                return;
            }
            
            if (url) {
                // Store the redirect URL instead of redirecting immediately
                setRedirectURL(url);
                setSocialIdpName(idpName);
            }
        }

        setLoading(false);
    }, [needsDecision, isSignupMode, clearToken, setToken, handleSocialLoginClick]);

    // Handle when user selects an authentication option
    const handleAuthOptionSelection = (actionId: string) => {
        setLoading(true);
        setSelectedAction(actionId);
        
        submitAuthDecision(flowId, actionId)
            .then((result) => {
                processAuthResponse(result.data);
            })
            .catch((error) => {
                console.error("Error during authentication decision:", error);
                setError(true);
                setErrorMessage(error.message || 'Error processing your selection');
                setLoading(false);
            });
    };

    const init = useCallback((isSignupMode: boolean = false) => {
        clearToken();
        setConnectionError(false);
        setNeedsDecision(false);
        setAvailableActions([]);
        setSelectedAction(null);
        setFormData({});
        setInputs([]);
        // Reset redirect URL
        setRedirectURL(null);
        setSocialIdpName('');
        setRegOnlySuccess(false);

        initiateNativeAuthFlow(isSignupMode ? 'REGISTRATION' : 'LOGIN')
            .then((result) => {
                const data = result.data;

                if (data.flowStatus && data.flowStatus === 'COMPLETE' && data.assertion) {
                    setToken(data.assertion);
                    setError(false);
                } else if (data.flowStatus && data.flowStatus === 'ERROR') {
                    const defaultMessage = isSignupMode 
                        ? 'Registration failed. Please check your information.' 
                        : 'Login failed. Please check your credentials.';
                    setError(true);
                    setErrorMessage(data.failureReason || defaultMessage);
                } else if (data.type === "VIEW") {
                    // Handle the VIEW response
                    if (data.data?.actions) {
                        // This is a decision screen
                        setNeedsDecision(true);
                        setAvailableActions(data.data.actions);
                    } else if (data.data?.inputs) {
                        // This is an input prompt
                        setNeedsDecision(false);
                        data.data.inputs.forEach((input: AuthInput) => {
                            setInputs(prev => [...prev, input]);
                        });
                    }
                } else if (data.type === "REDIRECTION") {
                    // Handle redirection for social logins
                    const url = data.data?.redirectURL;
                    const idpName = data.data?.additionalData?.idpName || 'Social Login';
                    
                    if (url) {
                        // Store the redirect URL instead of redirecting immediately
                        setRedirectURL(url);
                        setSocialIdpName(idpName);
                    }
                }

                setFlowId(data.flowId);
                setLoading(false);
            }).catch((error) => {
                const errorType = isSignupMode ? "registration" : "auth";
                console.error(`Error during ${errorType} initialization:`, error);
                setConnectionError(true);
                setLoading(false);
            });
    }, [clearToken, isSignupMode, setToken]);

    // Initialize the prompt signup decision action
    const initPromptSignupDecision = () => {
        clearToken();
        setConnectionError(false);
        setNeedsDecision(false);
        setAvailableActions([]);
        // Reset redirect URL
        setRedirectURL(null);
        setSocialIdpName('');
        setRegOnlySuccess(false);
        setPromptRegistration(false);

        // Ensure all input fields are present in formData, even if empty
        const completeFormData = { ...formData };
        inputs.forEach(input => {
            if (!(input.name in completeFormData)) {
                completeFormData[input.name] = '';
            }
        });

        initiateNativeAuthFlowWithData('REGISTRATION', selectedAction, completeFormData)
            .then((result) => {
                setInputs([]);
                const data = result.data;

                if (data.flowStatus && data.flowStatus === 'COMPLETE' && data.assertion) {
                    setToken(data.assertion);
                    setError(false);
                } else if (data.flowStatus && data.flowStatus === 'ERROR') {
                    setError(true);
                    setErrorMessage(data.failureReason || 'Registration failed. Please check your information.');
                } else if (data.type === "VIEW") {
                    // Handle the VIEW response
                    if (data.data?.actions) {
                        // This is a decision screen
                        setNeedsDecision(true);
                        setAvailableActions(data.data.actions);
                    } else if (data.data?.inputs) {
                        // This is an input prompt
                        setNeedsDecision(false);
                        data.data.inputs.forEach((input: AuthInput) => {
                            setInputs(prev => [...prev, input]);
                        });
                    }
                } else if (data.type === "REDIRECTION") {
                    // Handle redirection for social logins
                    const url = data.data?.redirectURL;
                    const idpName = data.data?.additionalData?.idpName || 'Social Login';
                    
                    if (url) {
                        // Store the redirect URL instead of redirecting immediately
                        setRedirectURL(url);
                        setSocialIdpName(idpName);
                    }
                }

                setFlowId(data.flowId);
                setLoading(false);
            }).catch((error) => {
                console.error(`Error during user registration:`, error);
                setInputs([]);
                setConnectionError(true);
                setLoading(false);
            });
    };

    // Unified form submission handler that works for both decisions and direct inputs
    const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        setLoading(true);

        // Ensure all input fields are present in formData, even if empty
        const completeFormData = { ...formData };
        inputs.forEach(input => {
            if (!(input.name in completeFormData)) {
                completeFormData[input.name] = '';
            }
        });

        const isMobileInput = inputs.some(input => input.name === "mobileNumber");

        if (needsDecision) {
            // This is a decision submission - identify the action from form data
            const formAction = event.currentTarget.getAttribute('data-action-id');
            if (formAction) {
                setSelectedAction(formAction);
                submitAuthDecision(flowId, formAction, completeFormData)
                    .then((result) => {
                        processAuthResponse(result.data, formAction);
                    })
                    .catch((error) => {
                        console.error("Error during authentication decision:", error);
                        handleSubmissionError(error);
                    });
            }
        } else {
            // This is a direct input submission
            submitNativeAuth(flowId, completeFormData)
                .then((result) => {
                    if (isMobileInput) {
                        processAuthResponse(result.data, "mobile");
                    } else {
                        processAuthResponse(result.data);
                    }
                })
                .catch((error) => {
                    console.error("Error during authentication:", error);
                    handleSubmissionError(error);
                });
        }
    };

    const handleSubmissionError = (error: SubmissionError) => {
        // Check if it's a network error or authentication error
        if (error.message && error.message.includes("Network Error")) {
            setConnectionError(true);
        } else {
            setError(true);
            setErrorMessage(error.message || 'Error during authentication');
        }
        setLoading(false);
    };

    const handleRetry = () => {
        setTimeout(() => {
            init();
        }, 500);
    };
    
    // Helper function to get appropriate icon for social login
    const getSocialLoginIcon = (idpName: string) => {
        const lowerIdpName = idpName.toLowerCase();
        
        if (lowerIdpName.includes('github')) {
            return <GitHubIcon />;
        } else if (lowerIdpName.includes('google')) {
            return <GoogleIcon />;
        } else {
            return <AccountCircleIcon />;
        }
    };

    // Get social login button text
    const getSocialLoginText = (actionId: string) => {
        const prefix = isSignupMode ? 'Sign up' : 'Continue';
        
        if (actionId.includes('google')) {
            return `${prefix} with Google`;
        } else if (actionId.includes('github')) {
            return `${prefix} with GitHub`;
        } else if (actionId.includes('mobile')) {
            return `${prefix} with SMS OTP`;
        } else {
            const idpText = actionId.split('_').map(word => 
                word.charAt(0).toUpperCase() + word.slice(1)
            ).join(' ');
            return `${prefix} with ${idpText}`;
        }
    };

    // This effect is to handle initial component mount
    useEffect(() => {
        // Prevent double mount due to React Strict Mode
        if (isComponentReMount.current) return;
        isComponentReMount.current = true;

        if (startInit) {
            // Initialize login execution flow if fresh start
            init();
        } else {
            // This effect is to handle when return from federated IDP login
            const params = new URLSearchParams(window.location.search);
            const code = params.get('code');

            if (code) {
                // Clear query parameters to avoid re-submission
                window.history.replaceState({}, document.title, window.location.pathname);

                submitNativeAuth(flowId, { type: NativeAuthSubmitType.SOCIAL, code: code })
                    .then((result) => {
                        processAuthResponse(result.data);
                    }).catch((error) => {
                        console.error("Error during social authentication:", error);
                        handleSubmissionError(error);
                    });
            } else {
                setError(true);
                setLoading(false);
            }

            sessionStorage.setItem(START_INIT_KEY, "true");
        }
    },[startInit, init, flowId, setToken, processAuthResponse]);

    // Render input fields based on the current inputs array
    const renderInputFields = () => {
        return inputs.map((input, index) => {
            const inputId = input.name || `input-${index}`;
            const isPassword = input.type === "password" || input.name === "password";
            const isOTP = input.type === "otp" || input.name === "otp";
            const isRequired = input.required;
            
            // Determine appropriate label
            let label = input.name;
            if (label) {
                label = label.charAt(0).toUpperCase() + label.slice(1).replace(/_/g, ' ');
            }
            label = label.replace(/([a-z])([A-Z])/g, '$1 $2');
            if (isOTP) {
                label = 'OTP Code';
            } else if (isPassword) {
                label = 'Password';
            }

            const placeholder = `Enter your ${label.toLowerCase()}`;

            if (isPassword) {
                return (
                    <Box key={inputId} display="flex" flexDirection="column" gap={0.5}>
                        <InputLabel htmlFor={inputId} sx={{ mb: 1 }}>{label}</InputLabel>
                        <OutlinedInput
                            type={showPassword ? 'text' : 'password'}
                            id={inputId}
                            name={input.name}
                            placeholder={placeholder}
                            size="small"
                            value={formData[input.name] || ''}
                            onChange={handleInputChange}
                            required={isRequired}
                            endAdornment={
                                <InputAdornment position="end">
                                    <IconButton
                                        aria-label="toggle password visibility"
                                        onClick={handleTogglePasswordVisibility}
                                        onMouseDown={handleMouseDownPassword}
                                        edge="end"
                                    >
                                        {showPassword ? <VisibilityOff /> : <Visibility />}
                                    </IconButton>
                                </InputAdornment>
                            }
                        />
                    </Box>
                );
            } else if (isOTP) {
                return (
                    <Box key={inputId} display="flex" flexDirection="column" gap={0.5}>
                        <InputLabel htmlFor={inputId} sx={{ mb: 1 }}>{label}</InputLabel>
                        <Box 
                            sx={{ 
                                display: 'flex', 
                                gap: 1,
                                justifyContent: 'space-between'
                            }}
                        >
                            {otpDigits.map((digit, index) => (
                                <OutlinedInput
                                    key={`otp-digit-${index}`}
                                    inputRef={el => otpInputRefs.current[index] = el}
                                    value={digit}
                                    onChange={(e) => handleOTPDigitChange(index, e.target.value)}
                                    onKeyDown={(e) => handleOTPKeyDown(index, e as React.KeyboardEvent<HTMLInputElement>)}
                                    onPaste={index === 0 ? handlePaste : undefined}
                                    inputProps={{
                                        maxLength: 1,
                                        style: { textAlign: 'center', padding: '8px 0' }
                                    }}
                                    sx={{
                                        width: '40px',
                                        height: '48px',
                                        '& input': { padding: 0 }
                                    }}
                                />
                            ))}
                        </Box>
                    </Box>
                );
            } else {
                return (
                    <Box key={inputId} display="flex" flexDirection="column" gap={0.5}>
                        <InputLabel htmlFor={inputId} sx={{ mb: 1 }}>{label}</InputLabel>
                        <OutlinedInput
                            type={input.type || "text"}
                            id={inputId}
                            name={input.name}
                            placeholder={placeholder}
                            size="small"
                            value={formData[input.name] || ''}
                            onChange={handleInputChange}
                            required={isRequired}
                        />
                    </Box>
                );
            }
        });
    };

    // Render the login form with side-by-side layout based on the available actions
    const renderSideBySideLoginForm = () => {
        const basicAuthAction = availableActions.find(action => action.id === "basic_auth");
        const mobileAuthActions = availableActions.filter(action => 
            action.id === "mobile_prompt_username" || action.id === "prompt_mobile"
        );
        
        const hasSocialAuth = availableActions.some(action => 
            action.id.includes("google") || action.id.includes("github")
        );
        const hasMobileAuth = mobileAuthActions.length > 0;
        
        const socialAuthActions = availableActions.filter(action => 
            action.id.includes("google") || action.id.includes("github")
        );
        
        return (
            <Box sx={{ my: 4 }}>
                <Box display="flex" gap={4}>
                    {/* Left: Basic Login */}
                    <Box sx={{ flex: 1 }}>
                        <form onSubmit={handleSubmit} data-action-id={basicAuthAction?.id}>
                            <Box display="flex" flexDirection="column" gap={2}  sx={{ mb: 2, mt: 6.8 }}>
                            </Box>
                            <Box display="flex" flexDirection="column" gap={2} sx={{ mt: 3 }}>
                                <Box display="flex" flexDirection="column" gap={0.5}>
                                    <InputLabel htmlFor="username">Username</InputLabel>
                                    <OutlinedInput
                                        type="text"
                                        id="username"
                                        name="username"
                                        placeholder="Enter your username"
                                        size="small"
                                        value={formData.username || ''}
                                        onChange={handleInputChange}
                                        required
                                    />
                                </Box>
                                <Box display="flex" flexDirection="column" gap={0.5} sx={{ mt: 1 }}>
                                    <InputLabel htmlFor="password">Password</InputLabel>
                                    <OutlinedInput
                                        type={showPassword ? 'text' : 'password'}
                                        id="password"
                                        name="password"
                                        placeholder="Enter your password"
                                        size="small"
                                        value={formData.password || ''}
                                        onChange={handleInputChange}
                                        required
                                        endAdornment={
                                        <InputAdornment position="end">
                                            <IconButton
                                                aria-label="toggle password visibility"
                                                onClick={handleTogglePasswordVisibility}
                                                onMouseDown={handleMouseDownPassword}
                                                edge="end"
                                            >
                                                {showPassword ? <VisibilityOff /> : <Visibility />}
                                            </IconButton>
                                        </InputAdornment>
                                        }
                                    />
                                </Box>

                                {(showRememberMe || showForgotPassword) && (
                                <Box display="flex" justifyContent="space-between" alignItems="center">
                                    {showRememberMe && (
                                        <FormControlLabel
                                            control={<Checkbox name="remember-me-checkbox" />}
                                            label="Remember me"
                                        />
                                    )}
                                    {showForgotPassword && <Link href="#">Forgot your password?</Link>}
                                </Box>
                                )}

                                <Button
                                    variant="contained"
                                    color="primary"
                                    type="submit"
                                    fullWidth
                                    sx={{ mt: 3 }}
                                    >
                                    {isSignupMode ? 'Create Account' : 'Sign In'}
                                </Button>
                            </Box>
                        </form>
                    </Box>

                    {/* Vertical Divider */}
                    <Divider orientation="vertical" flexItem sx={{ mx: 2 }} />

                    {/* Right: Social Auth and SMS Options */}
                    <Box sx={{ flex: 1 }}>
                        {/* Social auth options */}
                        {hasSocialAuth && (
                            <Box>
                                {socialAuthActions.map((action, index) => (
                                    <Button
                                        key={`social-action-${index}`}
                                        fullWidth
                                        variant="contained"
                                        color="secondary"
                                        onClick={() => handleAuthOptionSelection(action.id)}
                                        sx={{ my: 1 }}
                                        startIcon={getSocialLoginIcon(action.id)}
                                    >
                                        {getSocialLoginText(action.id)}
                                    </Button>
                                ))}
                            </Box>
                        )}

                        {/* Show divider if we have both social and sms auth options */}
                        {hasMobileAuth && hasSocialAuth && (
                            <Divider sx={{ my: 3 }}>or</Divider>
                        )}

                        {/* SMS OTP Auth */}
                        {hasMobileAuth && (
                            <form 
                                onSubmit={handleSubmit}
                                data-action-id={mobileAuthActions[0]?.id}
                            >
                                <Box display="flex" flexDirection="column" gap={2}>
                                    <Box display="flex" flexDirection="column" gap={0.5}>
                                        <InputLabel htmlFor={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}>
                                            {mobileAuthActions[0]?.id === "prompt_mobile" ? "Mobile Number" : "Username"}
                                        </InputLabel>
                                        <OutlinedInput
                                            type="text"
                                            id={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}
                                            name={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}
                                            placeholder={`Enter your ${mobileAuthActions[0]?.id === "prompt_mobile" ? "mobile number" : "username"}`}
                                            size="small"
                                            value={formData[mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"] || ''}
                                            onChange={handleInputChange}
                                            required
                                        />
                                    </Box>
                                    <Button
                                        variant="contained"
                                        color="primary"
                                        type="submit"
                                        fullWidth
                                        sx={{ mt: 2 }}
                                    >
                                        Continue with SMS OTP
                                    </Button>
                                </Box>
                            </form>
                        )}
                    </Box>
                </Box>
            </Box>
        );
    }

    // Render the regular login form with options stacked vertically
    const renderRegularLoginForm = () => {
        const basicAuthAction = availableActions.find(action => action.id === "basic_auth");
        const mobileAuthActions = availableActions.filter(action => 
            action.id === "mobile_prompt_username" || action.id === "prompt_mobile"
        );
        
        const hasBasicAuth = !!basicAuthAction;
        const hasSocialAuth = availableActions.some(action => 
            action.id.includes("google") || action.id.includes("github")
        );
        const hasMobileAuth = mobileAuthActions.length > 0;
        
        const socialAuthActions = availableActions.filter(action => 
            action.id.includes("google") || action.id.includes("github")
        );

        return (
            <Box sx={{ my: 2 }}>
                {/* Social auth options */}
                {hasSocialAuth && (
                    <Box>
                        {socialAuthActions.map((action, index) => (
                            <Button
                                key={`social-action-${index}`}
                                fullWidth
                                variant="contained"
                                color="secondary"
                                onClick={() => handleAuthOptionSelection(action.id)}
                                sx={{ my: 1 }}
                                startIcon={getSocialLoginIcon(action.id)}
                            >
                                {getSocialLoginText(action.id)}
                            </Button>
                        ))}
                    </Box>
                )}
                
                {/* Show divider if we have multiple auth options */}
                {((hasSocialAuth && hasBasicAuth) || (hasSocialAuth && hasMobileAuth)) && (
                    <Divider sx={{ my: 3 }}>or</Divider>
                )}
                
                {/* Basic auth form */}
                {hasBasicAuth && (
                    <form onSubmit={handleSubmit} data-action-id={basicAuthAction?.id}>
                        <Box display="flex" flexDirection="column" gap={2}>
                            <Box display="flex" flexDirection="column" gap={0.5}>
                                <InputLabel htmlFor="username">Username</InputLabel>
                                <OutlinedInput
                                    type="text"
                                    id="username"
                                    name="username"
                                    placeholder="Enter your username"
                                    size="small"
                                    value={formData.username || ''}
                                    onChange={handleInputChange}
                                    required
                                />
                            </Box>
                            <Box display="flex" flexDirection="column" gap={0.5}>
                                <InputLabel htmlFor="password">Password</InputLabel>
                                <OutlinedInput
                                    type={showPassword ? 'text' : 'password'}
                                    id="password"
                                    name="password"
                                    placeholder="Enter your password"
                                    size="small"
                                    value={formData.password || ''}
                                    onChange={handleInputChange}
                                    required
                                    endAdornment={
                                        <InputAdornment position="end">
                                            <IconButton
                                                aria-label="toggle password visibility"
                                                onClick={handleTogglePasswordVisibility}
                                                onMouseDown={handleMouseDownPassword}
                                                edge="end"
                                            >
                                                {showPassword ? 
                                                    <VisibilityOff /> : <Visibility />
                                                }
                                            </IconButton>
                                        </InputAdornment>
                                    }
                                />
                            </Box>
                            {(showRememberMe || showForgotPassword) && (
                                <Box
                                    sx={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    }}
                                >
                                    { showRememberMe && (
                                        <FormControlLabel
                                            control={<Checkbox name="remember-me-checkbox" />} 
                                            label="Remember me" />
                                    )}
                                    { showForgotPassword &&
                                        <Link href="">Forgot your password?</Link>
                                    }
                                </Box>
                            )}
                            <Button
                                variant="contained"
                                color="primary"
                                type="submit"
                                fullWidth
                                sx={{ mt: 2 }}
                            >
                                {isSignupMode ? 'Create Account' : 'Sign In'}
                            </Button>
                        </Box>
                    </form>
                )}

                {/* Show divider if we have multiple auth options */}
                {(hasBasicAuth && hasMobileAuth) && (
                    <Divider sx={{ my: 3 }}>or</Divider>
                )}

                {/* SMS OTP auth form */}
                {hasMobileAuth && (
                    <form 
                        onSubmit={handleSubmit} 
                        data-action-id={mobileAuthActions[0]?.id}
                    >
                        <Box display="flex" flexDirection="column" gap={2}>
                            <Box display="flex" flexDirection="column" gap={0.5}>
                                <InputLabel htmlFor={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}>
                                    {mobileAuthActions[0]?.id === "prompt_mobile" ? "Mobile Number" : "Username"}
                                </InputLabel>
                                <OutlinedInput
                                    type="text"
                                    id={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}
                                    name={mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"}
                                    placeholder={`Enter your ${mobileAuthActions[0]?.id === "prompt_mobile" ? "mobile number" : "username"}`}
                                    size="small"
                                    value={formData[mobileAuthActions[0]?.id === "prompt_mobile" ? "mobileNumber" : "username"] || ''}
                                    onChange={handleInputChange}
                                    required
                                />
                            </Box>
                            <Button
                                variant="contained"
                                color="primary"
                                type="submit"
                                fullWidth
                                sx={{ mt: 2 }}
                            >
                                Continue with SMS OTP
                            </Button>
                        </Box>
                    </form>
                )}
            </Box>
        );
    }

    const renderInputPromptForm = () => {
        return (
            <form onSubmit={handleSubmit}>
                <Box display="flex" flexDirection="column" gap={2}>
                    {renderInputFields()}
                    
                    {(showRememberMe || showForgotPassword) && (
                        <Box
                            sx={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            }}
                        >
                            { showRememberMe && (
                                <FormControlLabel
                                    control={<Checkbox name="remember-me-checkbox" />} 
                                    label="Remember me" />
                            )}
                            { showForgotPassword &&
                                <Link href="">Forgot your password?</Link>
                            }
                        </Box>
                    )}

                    { error && !errorMessage.includes("invalid OTP") ? (
                        <Button
                            variant="contained"
                            color="primary"
                            type="submit"
                            fullWidth
                            sx={{ mt: 2 }}
                            onClick={(e) => {
                                e.preventDefault();
                                setError(false);
                                setErrorMessage('');
                                handleRetry();
                            }}
                         >
                            Retry
                        </Button>
                    ) : (
                        <Button
                            variant="contained"
                            color="primary"
                            type="submit"
                            fullWidth
                            sx={{ mt: 2 }}
                        >
                            {
                                inputs.some(input => input.name === 'password') ? 
                                    (isSignupMode ? 
                                        'Create Account' 
                                        : 'Sign In'
                                    ) 
                                    : inputs.some(input => input.name === 'otp') ? 
                                        'Verify OTP' 
                                        : 'Continue'
                            }
                        </Button>
                    )}
                </Box>
            </form>
        );
    }

    // Render function for redirection scenarios
    const renderRedirectLoginButton = () => {
        if (!redirectURL) return null;
        
        const buttonText = socialIdpName ? 
            `Continue with ${socialIdpName}` :
            'Continue with Social Login';
        
        const icon = getSocialLoginIcon(socialIdpName);
        
        return (
            <Box sx={{ my: 2 }}>
                <Button
                    fullWidth
                    variant="contained"
                    color="secondary"
                    onClick={() => handleSocialLoginClick(redirectURL)}
                    sx={{ my: 1 }}
                    startIcon={icon}
                >
                    {buttonText}
                </Button>
            </Box>
        );
    };

    // Calculate appropriate grid size based on layout complexity
    const gridMdSize = needsDecision 
        && !promptRegistration
        && availableActions.some(action => action.id === "basic_auth") 
        && availableActions.some(action => action.id === "mobile_prompt_username" || action.id === "prompt_mobile") 
        ? 10 : 6;
    const containerBoxMaxWidth = gridMdSize === 10 ? 1000 : 500;

    const basicAuthAction = availableActions.find(action => action.id === "basic_auth");
    const mobileAuthActions = availableActions.filter(action => 
        action.id === "mobile_prompt_username" || action.id === "prompt_mobile"
    );
    
    const hasBasicAuth = !!basicAuthAction;
    const hasMobileAuth = mobileAuthActions.length > 0;

    return (
        <Layout>
            { loading ? (
                <GradientCircularProgress />
            ) : (
                <Grid size={{ xs: 12, md: gridMdSize }}>
                    <Paper
                        sx={{
                            display: "flex",
                            width: "100%",
                            height: "100%",
                            flexDirection: "column",
                        }}
                    >
                        <Box
                            sx={{
                                alignItems: "center",
                                justifyContent: "center",
                                padding: 6,
                                width: "100%",
                                maxWidth: containerBoxMaxWidth,
                                margin: "auto",
                            }}
                        >
                            <Box>
                                {promptRegistration ? (
                                    <Box sx={{ mb: 4 }}>
                                        <Typography variant="h5" gutterBottom>
                                            We couldn&apos;t find your account
                                        </Typography>
                                        <Typography>
                                            No account matched your details. You can try again or sign up below.
                                        </Typography>
                                    </Box>
                                ) : regOnlySuccess ? (
                                    <Box sx={{ mb: 4 }}>
                                        <Typography variant="h5" gutterBottom>
                                            Registration Successful
                                        </Typography>
                                        <Typography>
                                            You can now log in to your account.
                                        </Typography>
                                    </Box>
                                ) : (
                                    <Box sx={{ mb: 4 }}>
                                        <Typography variant="h5" gutterBottom>
                                            {isSignupMode ? 'Create Account' : 'Login to Account'}
                                        </Typography>

                                        <Typography>
                                            {isSignupMode ? (
                                                <>
                                                    Already have an account?{' '}
                                                    <Link 
                                                        href="#" 
                                                        onClick={(e) => {
                                                            e.preventDefault();
                                                            setIsSignupMode(false);
                                                            setError(false);
                                                            setErrorMessage('');
                                                            init(false);
                                                        }}
                                                        underline="hover"
                                                    >
                                                        Sign in!
                                                    </Link>
                                                </>
                                            ) : (
                                                <>
                                                    Don&apos;t have an account?{' '}
                                                    <Link 
                                                        href="#" 
                                                        onClick={(e) => {
                                                            e.preventDefault();
                                                            setIsSignupMode(true);
                                                            setError(false);
                                                            setErrorMessage('');
                                                            init(true);
                                                        }}
                                                        underline="hover"
                                                    >
                                                        Sign up!
                                                    </Link>
                                                </>
                                            )}
                                        </Typography>
                                    </Box>
                                )}
                                
                                {connectionError && (
                                    <ConnectionErrorModal 
                                        onRetry={handleRetry}
                                        retryCount={retryCount}
                                        onRetryCountIncrement={() => setRetryCount(prev => prev + 1)}
                                    />
                                )}

                                {error && !connectionError && (
                                    <Alert severity="error" sx={{ my: 2 }}>
                                        {errorMessage}
                                    </Alert>
                                )}

                                {!connectionError && (
                                    promptRegistration ? (
                                        <Box sx={{ mb: 4 }}>
                                            <Button
                                                variant="contained"
                                                color="primary"
                                                type="submit"
                                                fullWidth
                                                sx={{ mt: 2 }}
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    setIsSignupMode(true);
                                                    setError(false);
                                                    setErrorMessage('');
                                                    initPromptSignupDecision();
                                                }}
                                            >
                                                Sign Up
                                            </Button>
                                            <Button
                                                variant="contained"
                                                color="secondary"
                                                type="submit"
                                                fullWidth
                                                sx={{ mt: 2 }}
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    setPromptRegistration(false);
                                                    setError(false);
                                                    setErrorMessage('');
                                                    handleRetry();
                                                }}
                                            >
                                                Retry
                                            </Button>
                                        </Box>
                                    ) : !regOnlySuccess ? (
                                    <>
                                        {/* First check if we have a redirect URL */}
                                        {redirectURL ? (
                                            renderRedirectLoginButton()
                                        ) : needsDecision ? (
                                            /* If not redirect but needs decision */
                                            <>
                                                { hasBasicAuth && hasMobileAuth ? (
                                                    renderSideBySideLoginForm()
                                                ) : (
                                                    renderRegularLoginForm()
                                                )}
                                            </>
                                        ) : (
                                            /* If not redirect and not decision, it's an input prompt */
                                            renderInputPromptForm()
                                        )}
                                    </>
                                    ) : (
                                        <Button
                                            variant="contained"
                                            color="primary"
                                            onClick={() => {
                                                setIsSignupMode(false);
                                                setError(false);
                                                setErrorMessage('');
                                                init(false);
                                            }}
                                            fullWidth
                                        >
                                            Go to Login
                                        </Button>
                                    )
                                )}
                                
                                <Box component="footer" sx={{ mt: 6 }}>
                                    <Typography sx={{ textAlign: "center" }}>
                                        © Copyright {new Date().getFullYear()}
                                    </Typography>
                                </Box>
                            </Box>
                        </Box>
                    </Paper>
                </Grid>
            )}
        </Layout>
    );
};

export default LoginPage;
