// m/frontend/src/pages/Recommendations.jsx
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container, Grid, Card, CardContent, CardMedia, Typography, Button, CardActions,
CircularProgress, ToggleButton, ToggleButtonGroup, Box, TextField, FormControl, InputLabel, 
Select, MenuItem, Checkbox, ListItemText, FormControlLabel, CardActionArea } from '@mui/material';
import { toast } from 'react-toastify';
import { getRecommendations, declineRecommendation } from '../api/recommendations';
import { getUser, getUserBio } from '../api/user';
import { sendConnectionRequest } from '../api/connections';
import { getConnections, getPendingConnections } from '../api/connections';
import { getSentConnections } from '../api/connections';
import axios from 'axios';
const cityOptions = [
  { name: 'Helsinki', lat: 60.1699, lon: 24.9384 },
  { name: 'Espoo', lat: 60.2055, lon: 24.6559 },
  { name: 'Vantaa', lat: 60.2934, lon: 25.0378 },
  { name: 'Turku', lat: 60.4518, lon: 22.2666 },
  { name: 'Tampere', lat: 61.4981, lon: 23.7610 },
  { name: 'Oulu', lat: 65.0121, lon: 25.4651 },
  { name: 'Lahti', lat: 60.9827, lon: 25.6615 },
  { name: 'Kuopio', lat: 62.8924, lon: 27.6770 },
  { name: 'Pori', lat: 61.4850, lon: 21.7973 },
  { name: 'Jyväskylä', lat: 62.2426, lon: 25.7473 },
];
const interestsOptions = ["movies", "sports", "music", "technology", "art"];
const hobbiesOptions   = ["reading", "running", "drawing", "gaming", "cooking"];
const musicOptions     = ["rock", "jazz", "classical", "pop", "hip-hop"];
const foodOptions      = ["italian", "asian", "russian", "french", "mexican"];
const travelOptions    = ["beach", "mountains", "cities", "expeditions", "ecotourism"];
/**
 * Recommendations.jsx
 * 
 * Main recommendations page. Handles user matching logic, filtering, and connection requests.
 * Integrates with backend API for recommendations, user info, and connections.
 * Supports two modes: affinity (profile-based) and desire (custom search).
 * Handles batch loading, filtering, and error-driven navigation.
 */
const Recommendations = () => {
  const navigate = useNavigate();
  const [mode, setMode] = useState('affinity');
  const [useProfileFilters, setUseProfileFilters] = useState(true);
  const [connections, setConnections] = useState([]);
  const [pending, setPending]       = useState([]);
  const [form, setForm] = useState({
    city: cityOptions[0],
    interests: [], priorityInterests: false,
    hobbies:   [], priorityHobbies: false,
    music:     [], priorityMusic: false,
    food:      [], priorityFood: false,
    travel:    [], priorityTravel: false,
    lookingFor: ''
  });
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(false);
  const [decliningId, setDecliningId] = useState(null);
  const [radius, setRadius] = useState(null);

useEffect(() => {
  // Batch-load connections and pending requests on mount
  Promise.all([getConnections(), getPendingConnections()])
    .then(([conns, pend]) => {
      setConnections(conns);
      setPending(pend);
    });
}, []);

const [sent, setSent] = useState([]);

useEffect(() => {
  // Batch-load pending and sent requests for up-to-date state
  Promise.all([
    getPendingConnections(),
    getSentConnections()
  ]).then(([pend, sent]) => {
    setPending(pend);
    setSent(sent);
  });
}, []);

const fetchLinks = async () => {
  // Loads all connection-related lists in parallel, updates state, handles errors
  try {
    const [conns, pend, sent] = await Promise.all([
      getConnections(),
      getPendingConnections(),
      getSentConnections(),
    ]);
    setConnections(conns);
    setPending(pend);
    setSent(sent);
    } catch {
      toast.error('Failed to load connections');
    }
  };
  
  useEffect(() => {
    fetchLinks();
  }, []);

useEffect(() => {
  // Get maxRadius from user settings
  axios.get('/me/preferences')
    .then(res => setRadius(res.data.maxRadius))
    .catch(() => setRadius(null));
}, []);

  /**
   * handleSearch
   * Submits the recommendation search form. Builds query params based on mode and filters.
   * Handles API errors, including profile validation errors (redirects to edit profile).
   * Updates recommendations state with filtered results.
   */
  const handleSearch = async e => {
    e.preventDefault();
    await fetchLinks();
    setLoading(true);
     const params = { mode, withDistance: true, useProfile: useProfileFilters };
    if (!useProfileFilters) {
      params.cityLat = form.city.lat;
      params.cityLon = form.city.lon;
      if (mode === 'affinity') {
        params.interests         = form.interests.join(',');
        params.priorityInterests = form.priorityInterests;
        params.hobbies           = form.hobbies.join(',');
        params.priorityHobbies   = form.priorityHobbies;
        params.music             = form.music.join(',');
        params.priorityMusic     = form.priorityMusic;
        params.food              = form.food.join(',');
        params.priorityFood      = form.priorityFood;
        params.travel            = form.travel.join(',');
        params.priorityTravel    = form.priorityTravel;
      } else {
        params.lookingFor = form.lookingFor;
      }
    }
    try {
      const recs = await getRecommendations({ params: { ...params, limit: 20 } });
      const recData = await Promise.all(
        recs.map(async ({ id,  distance, score  }) => {
          try {
            const user = await getUser(id);
            const bio  = await getUserBio(id);
            return { id,  distance, score , ...user, bio };
          } catch (err) {
            console.error(`[ERROR] Failed to load user ${id}:`, err);
            return null;
          }
        })
      );

      const filtered = recData
        .filter(r => r && r.firstName && r.lastName)
        .filter(r => !connections.includes(r.id));

      setRecommendations(filtered.slice(0, 10));

    } catch (err) {
      const msg = err.response?.data || 'Error loading recommendations';
      toast.error(msg);
      if (/fill in/i.test(msg)) {
        setTimeout(() => navigate('/edit-profile'), 2000);
      }
    } finally {
      setLoading(false);
    }
  };
  /**
   * handleDecline
   * Declines a recommendation, updates UI optimistically, handles errors.
   */
  const handleDecline = async id => {
    setDecliningId(id);
    try {
      await declineRecommendation(id);
      setRecommendations(prev => prev.filter(r => r.id !== id));
      toast.success('Recommendation declined');
    } catch {
      toast.error('Error declining recommendation');
    } finally {
      setDecliningId(null);
    }
  };
  /**
   * handleConnect
   * Sends a connection request, updates pending state, handles errors.
   */
  const handleConnect = async id => {
    try {
      await sendConnectionRequest(id);
      toast.success('Request sent');
      setPending(prev => [...prev, id]);
    } catch {
      toast.error('Error sending request');
    }
  };
  
  /**
   * switchMode
   * Switches between affinity and desire modes, resets recommendations.
   */
  const switchMode = (newMode) => {
    if (newMode !== mode) {
      setMode(newMode);
      setRecommendations([]);
    }
  };
  return (
    <Container sx={{ mt: 4 }}>
      <Box sx={{ mb: 2, p: 2, background: '#fffbe6', border: '1px solid #ffe58f', borderRadius: 2 }}>
        <Typography variant="body1">
          Let's help you find great matches! ) Just set your search radius in <Button size="small" onClick={() => navigate('/settings')}>Settings</Button> to get started.
        </Typography>
      </Box>
      <Box sx={{ mb: 2, display: 'flex', alignItems: 'center', gap: 2 }}>
        <Button
          variant={mode === 'affinity' ? 'contained' : 'outlined'}
          onClick={() => switchMode('affinity')}
        >
          AffinityMatch
        </Button>
        <Button
          variant={mode === 'desire' ? 'contained' : 'outlined'}
          onClick={() => switchMode('desire')}
        >
          DesireMatch
        </Button>
        <FormControlLabel
          control={
            <Checkbox
              checked={useProfileFilters}
              onChange={e => setUseProfileFilters(e.target.checked)}
            />
          }
          label="Use profile data"
          sx={{ ml: 2 }}
        />
      </Box>
      <Typography variant="h4" gutterBottom>Recommendations</Typography>
      <Box component="form" onSubmit={handleSearch} sx={{ mb: 4 }}>
        <FormControl sx={{ minWidth: 200, mr: 2 }}
        disabled={useProfileFilters}>
          <InputLabel>City</InputLabel>
          <Select
            value={form.city.name}
            label="City"
            onChange={e => {
              const sel = cityOptions.find(c => c.name === e.target.value);
              setForm(f => ({ ...f, city: sel }));
            }}
          >
            {cityOptions.map(c => (
              <MenuItem key={c.name} value={c.name}>{c.name}</MenuItem>
            ))}
          </Select>
        </FormControl>
        {mode === 'affinity' ? (
          <>
            {[
              ['Interests', 'interests', interestsOptions, 'priorityInterests'],
              ['Hobbies', 'hobbies', hobbiesOptions, 'priorityHobbies'],
              ['Music', 'music', musicOptions, 'priorityMusic'],
              ['Food', 'food', foodOptions, 'priorityFood'],
              ['Travel', 'travel', travelOptions, 'priorityTravel']
            ].map(([label, key, opts, prioKey]) => (
              <FormControl key={key} sx={{ minWidth: 200, mr: 2, mt: 2 }}
              disabled={useProfileFilters}>
                <InputLabel>{label}</InputLabel>
                <Select
                  multiple
                  value={form[key]}
                  onChange={e => setForm(f => ({ ...f, [key]: e.target.value }))}
                  renderValue={selected => selected.join(', ')}
                  label={label}
                >
                  {opts.map(opt => (
                    <MenuItem key={opt} value={opt}>
                      <Checkbox checked={form[key].includes(opt)} 
                      disabled={useProfileFilters}/>
                      <ListItemText primary={opt} />
                    </MenuItem>
                  ))}
                </Select>
                <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                  <Checkbox
                    checked={form[prioKey]}
                    onChange={e => setForm(f => ({ ...f, [prioKey]: e.target.checked }))}
                    disabled={useProfileFilters}/>
                  <Typography variant="body2">Priority</Typography>
                </Box>
              </FormControl>
            ))}
          </>
        ) : (
          <TextField
            label="Who are you looking for"
            value={form.lookingFor}
            onChange={e => setForm(f => ({ ...f, lookingFor: e.target.value }))}
            disabled={useProfileFilters}
            sx={{ minWidth: 300, mr: 2 }}
          />
        )}
        <Button
          type="submit"
          variant="contained"
          sx={{ mt: 2 }}
          disabled={loading}
        >
          Match Me
        </Button>
      </Box>
      {loading ? (
        <Box sx={{ textAlign: 'center', mt: 4 }}>
          <CircularProgress />
        </Box>
      ) : recommendations.length === 0 ? (
        <Typography>No recommendations available.</Typography>
      ) : (
        <Grid container spacing={3}>
          {recommendations.map(rec => (
            <Grid item xs={12} sm={6} md={4} key={rec.id}>
              <Card>
              {/* <CardActionArea onClick={() => navigate(`/users/${rec.id}`)}> */}
              <CardActionArea
   onClick={() =>
     navigate(`/users/${rec.id}`, {
       state: { distance: rec.distance, score: rec.score }
     })
   }>
                <CardMedia
                  component="img"
                  height="140"
                  image={rec.photoUrl || '/static/images/default.png'}
                  alt={`${rec.firstName} ${rec.lastName}`}
                />
                <CardContent>
                  <Typography variant="h5">
                    {rec.firstName} {rec.lastName}
                  </Typography>
                  {/* {typeof rec.distance === 'number' && (
                    <Typography variant="body2" color="text.secondary">
                      Distance: {rec.distance.toFixed(1)} km
                    </Typography>
                  )}
                  {typeof rec.score === 'number' && (
                    <Typography variant="body2" color="text.secondary">
                      Match: {(rec.score * 100).toFixed(0)} %
                    </Typography>
                  )} */}
                  <Typography variant="body2" color="text.secondary">
                    {rec.bio.interests
                      ? `Interests: ${rec.bio.interests}`
                      : 'Information not available'}
                  </Typography>
                </CardContent>
                </CardActionArea>
                <CardActions>
                  {(() => {
                    const isFriend = connections.includes(rec.id);
                    const isPendingReq = pending.includes(rec.id) || sent.includes(rec.id);

                    if (isFriend) {
                      return (
                        <Button size="small" variant="contained" disabled>
                          Friends
                        </Button>
                      );
                    }
                    if (isPendingReq) {
                      return (
                        <Button size="small" variant="contained" disabled>
                          Request Sent
                        </Button>
                      );
                    }
                    return (
                      <Button
                        size="small"
                        variant="contained"
                        onClick={() => handleConnect(rec.id)}
                      >
                        Connect
                      </Button>
                    );
                  })()}

                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => handleDecline(rec.id)}
                    disabled={decliningId === rec.id}
                  >
                    {decliningId === rec.id ? 'Declining...' : 'Decline'}
                  </Button>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Container>
  );
};
export default Recommendations;
